package httphandler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/RomanKovalev007/url-shortner/internal/domain"
	httphandler "github.com/RomanKovalev007/url-shortner/internal/transport/http/handler"
	httptransport "github.com/RomanKovalev007/url-shortner/internal/transport/http"
)

type mockService struct {
	shorten     func(ctx context.Context, original string) (domain.URL, error)
	getOriginal func(ctx context.Context, alias string) (domain.URL, error)
}

func (m *mockService) Shorten(ctx context.Context, original string) (domain.URL, error) {
	return m.shorten(ctx, original)
}

func (m *mockService) GetOriginal(ctx context.Context, alias string) (domain.URL, error) {
	return m.getOriginal(ctx, alias)
}

func newRouter(svc *mockService) http.Handler {
	h := httphandler.NewHandler(svc, "http://short.ly")
	return httptransport.NewRouter(h)
}

func fixedURL(alias, original string) domain.URL {
	return domain.URL{ID: uuid.New(), Alias: alias, Original: original, CreatedAt: time.Now()}
}

func TestShorten_Success(t *testing.T) {
	router := newRouter(&mockService{
		shorten: func(_ context.Context, original string) (domain.URL, error) {
			return fixedURL("abc123_XYZ", original), nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusCreated)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["alias"] != "abc123_XYZ" {
		t.Errorf("got alias %q, want %q", resp["alias"], "abc123_XYZ")
	}
	if !strings.Contains(resp["short_url"], "abc123_XYZ") {
		t.Errorf("short_url %q does not contain alias", resp["short_url"])
	}
}

func TestShorten_InvalidJSON(t *testing.T) {
	router := newRouter(&mockService{})

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`not json`))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestShorten_EmptyURL(t *testing.T) {
	router := newRouter(&mockService{})

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":""}`))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestShorten_InvalidURLScheme(t *testing.T) {
	cases := []string{
		"ftp://example.com",
		"example.com",
		"//example.com",
	}

	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			router := newRouter(&mockService{})
			body := strings.NewReader(`{"url":"` + raw + `"}`)
			req := httptest.NewRequest(http.MethodPost, "/shorten", body)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("url %q: got status %d, want %d", raw, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestShorten_ServiceError(t *testing.T) {
	router := newRouter(&mockService{
		shorten: func(_ context.Context, _ string) (domain.URL, error) {
			return domain.URL{}, errors.New("internal error")
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestRedirectToOriginal_Success(t *testing.T) {
	router := newRouter(&mockService{
		getOriginal: func(_ context.Context, _ string) (domain.URL, error) {
			return fixedURL("abc123_XYZ", "https://example.com"), nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/abc123_XYZ", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusFound)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com" {
		t.Errorf("got Location %q, want %q", loc, "https://example.com")
	}
}

func TestRedirectToOriginal_NotFound(t *testing.T) {
	router := newRouter(&mockService{
		getOriginal: func(_ context.Context, _ string) (domain.URL, error) {
			return domain.URL{}, domain.ErrNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/unknownAls", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRedirectToOriginal_ServiceError(t *testing.T) {
	router := newRouter(&mockService{
		getOriginal: func(_ context.Context, _ string) (domain.URL, error) {
			return domain.URL{}, errors.New("internal error")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/abc123_XYZ", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHealth(t *testing.T) {
	router := newRouter(&mockService{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("got status %q, want %q", resp["status"], "ok")
	}
}
