package httphandler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/RomanKovalev007/url-shortner/internal/domain"
)

type urlService interface {
	Shorten(ctx context.Context, original string) (domain.URL, bool, error)
	GetOriginal(ctx context.Context, alias string) (domain.URL, error)
}

type Handler struct {
	service urlService
	baseURL string
}

func NewHandler(service urlService, baseURL string) *Handler {
	return &Handler{service: service, baseURL: baseURL}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validateURL(req.URL); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, created, err := h.service.Shorten(r.Context(), req.URL)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to shorten url", "error", err, "url", req.URL)
		writeError(w, http.StatusInternalServerError, "failed to shorten url")
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, shortenResponse{
		Alias:    res.Alias,
		ShortURL: h.baseURL + "/" + res.Alias,
	})
}

func (h *Handler) RedirectToOriginal(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")

	res, err := h.service.GetOriginal(r.Context(), alias)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "alias not found")
			return
		}
		slog.ErrorContext(r.Context(), "failed to resolve alias", "error", err, "alias", alias)
		writeError(w, http.StatusInternalServerError, "failed to resolve alias")
		return
	}

	http.Redirect(w, r, res.Original, http.StatusFound)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}


