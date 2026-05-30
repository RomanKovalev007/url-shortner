package httptransport

import (
	"net/http"

	"github.com/RomanKovalev007/url-shortner/internal/transport/http/middleware"
)

type httpHandler interface {
	RedirectToOriginal(w http.ResponseWriter, r *http.Request)
	Shorten(w http.ResponseWriter, r *http.Request)
	Health(w http.ResponseWriter, r *http.Request)
}

func NewRouter(h httpHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", h.Shorten)
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /{alias}", h.RedirectToOriginal)

	return middleware.Logging(mux)
}