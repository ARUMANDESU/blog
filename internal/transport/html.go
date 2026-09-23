package transport

import (
	"net/http"

	"github.com/arumandesu/blog/internal/views"
)

type HTTP struct {
}

func Handle(mux *http.ServeMux, h *HTTP) {
	mux.HandleFunc("GET /", h.GetHome)
	mux.HandleFunc("GET /post", h.GetPost)
}

func (h *HTTP) GetHome(w http.ResponseWriter, r *http.Request) {
	_ = views.Home().Render(r.Context(), w)
}

func (h *HTTP) GetPost(w http.ResponseWriter, r *http.Request) {
	_ = views.Post().Render(r.Context(), w)
}
