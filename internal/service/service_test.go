package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/RomanKovalev007/url-shortner/internal/domain"
	"github.com/RomanKovalev007/url-shortner/internal/service"
)

const aliasAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

type mockRepo struct {
	saveAlias  func(ctx context.Context, alias, original string) (domain.URL, error)
	getByAlias func(ctx context.Context, alias string) (domain.URL, error)
}

func (m *mockRepo) SaveAlias(ctx context.Context, alias, original string) (domain.URL, error) {
	return m.saveAlias(ctx, alias, original)
}

func (m *mockRepo) GetByAlias(ctx context.Context, alias string) (domain.URL, error) {
	return m.getByAlias(ctx, alias)
}

func fixedURL(alias, original string) domain.URL {
	return domain.URL{
		ID: uuid.New(), 
		Alias: alias, 
		Original: original, 
		CreatedAt: time.Now(),
	}
}

func TestShorten(t *testing.T) {
	ctx := context.Background()

	t.Run("returns url on success", func(t *testing.T) {
		svc := service.NewService(&mockRepo{
			saveAlias: func(_ context.Context, alias, original string) (domain.URL, error) {
				return fixedURL(alias, original), nil
			},
		})

		got, err := svc.Shorten(ctx, "https://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Original != "https://example.com" {
			t.Errorf("got original %q, want %q", got.Original, "https://example.com")
		}
		if len(got.Alias) != 10 {
			t.Errorf("alias length %d, want 10", len(got.Alias))
		}
	})

	t.Run("retries on ErrAliasAlreadyExists and succeeds", func(t *testing.T) {
		calls := 0
		svc := service.NewService(&mockRepo{
			saveAlias: func(_ context.Context, alias, original string) (domain.URL, error) {
				calls++
				if calls < 3 {
					return domain.URL{}, domain.ErrAliasAlreadyExists
				}
				return fixedURL(alias, original), nil
			},
		})

		_, err := svc.Shorten(ctx, "https://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if calls != 3 {
			t.Errorf("expected 3 calls, got %d", calls)
		}
	})

	t.Run("returns error after max retries exhausted", func(t *testing.T) {
		svc := service.NewService(&mockRepo{
			saveAlias: func(_ context.Context, _ string, _ string) (domain.URL, error) {
				return domain.URL{}, domain.ErrAliasAlreadyExists
			},
		})

		_, err := svc.Shorten(ctx, "https://example.com")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns unexpected repo error", func(t *testing.T) {
		repoErr := errors.New("db connection lost")
		svc := service.NewService(&mockRepo{
			saveAlias: func(_ context.Context, _ string, _ string) (domain.URL, error) {
				return domain.URL{}, repoErr
			},
		})

		_, err := svc.Shorten(ctx, "https://example.com")
		if !errors.Is(err, repoErr) {
			t.Errorf("got %v, want %v", err, repoErr)
		}
	})

	t.Run("returns existing url for duplicate original", func(t *testing.T) {
		existing := fixedURL("existAlias_", "https://example.com")
		svc := service.NewService(&mockRepo{
			saveAlias: func(_ context.Context, _ string, _ string) (domain.URL, error) {
				return existing, nil
			},
		})

		got, err := svc.Shorten(ctx, "https://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Alias != existing.Alias {
			t.Errorf("got alias %q, want %q", got.Alias, existing.Alias)
		}
	})

	t.Run("succeeds on last allowed retry", func(t *testing.T) {
		const maxRetries = 5
		calls := 0
		svc := service.NewService(&mockRepo{
			saveAlias: func(_ context.Context, alias, original string) (domain.URL, error) {
				calls++
				if calls < maxRetries {
					return domain.URL{}, domain.ErrAliasAlreadyExists
				}
				return fixedURL(alias, original), nil
			},
		})

		_, err := svc.Shorten(ctx, "https://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if calls != maxRetries {
			t.Errorf("expected %d calls, got %d", maxRetries, calls)
		}
	})

	t.Run("retries makes exactly maxRetries repo calls", func(t *testing.T) {
		const maxRetries = 5
		calls := 0
		svc := service.NewService(&mockRepo{
			saveAlias: func(_ context.Context, _ string, _ string) (domain.URL, error) {
				calls++
				return domain.URL{}, domain.ErrAliasAlreadyExists
			},
		})

		if _, err := svc.Shorten(ctx, "https://example.com"); err == nil {
			t.Fatal("expected error, got nil")
		}
		if calls != maxRetries {
			t.Errorf("expected exactly %d calls, got %d", maxRetries, calls)
		}
	})

	t.Run("generated alias contains only valid characters", func(t *testing.T) {
		svc := service.NewService(&mockRepo{
			saveAlias: func(_ context.Context, alias, original string) (domain.URL, error) {
				return fixedURL(alias, original), nil
			},
		})

		for range 50 {
			got, err := svc.Shorten(ctx, "https://example.com")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, ch := range got.Alias {
				if !strings.ContainsRune(aliasAlphabet, ch) {
					t.Errorf("alias %q contains invalid character %q", got.Alias, ch)
				}
			}
		}
	})
}

func TestGetOriginal(t *testing.T) {
	ctx := context.Background()

	t.Run("returns url for known alias", func(t *testing.T) {
		expected := fixedURL("knownAlias_", "https://example.com")
		svc := service.NewService(&mockRepo{
			getByAlias: func(_ context.Context, alias string) (domain.URL, error) {
				return expected, nil
			},
		})

		got, err := svc.GetOriginal(ctx, "knownAlias_")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Original != expected.Original {
			t.Errorf("got %q, want %q", got.Original, expected.Original)
		}
	})

	t.Run("returns ErrNotFound", func(t *testing.T) {
		svc := service.NewService(&mockRepo{
			getByAlias: func(_ context.Context, _ string) (domain.URL, error) {
				return domain.URL{}, domain.ErrNotFound
			},
		})

		_, err := svc.GetOriginal(ctx, "unknown")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("returns unexpected repo error", func(t *testing.T) {
		repoErr := errors.New("db connection lost")
		svc := service.NewService(&mockRepo{
			getByAlias: func(_ context.Context, _ string) (domain.URL, error) {
				return domain.URL{}, repoErr
			},
		})

		_, err := svc.GetOriginal(ctx, "someAlias_")
		if !errors.Is(err, repoErr) {
			t.Errorf("got %v, want %v", err, repoErr)
		}
	})
}
