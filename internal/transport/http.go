package transport

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"uuid"

	"github.com/arumandesu/blog/assets"
	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/domain"
	"github.com/arumandesu/blog/internal/views"
	"github.com/arumandesu/blog/pkg"
)

type HTTP struct {
	app    *app.App
	logger *slog.Logger
}

func NewHTTP(a *app.App, logger *slog.Logger) *HTTP {
	return &HTTP{app: a, logger: logger}
}

func Handle(mux *http.ServeMux, h *HTTP) {
	fileServer := http.FileServerFS(assets.StaticFiles)

	mux.Handle("GET /static/", cacheImmutable(http.StripPrefix("/static", fileServer), "/static/fonts/"))
	mux.HandleFunc("GET /", h.GetHome)
	mux.HandleFunc("GET /posts/{slug}", h.GetPost)
	mux.HandleFunc("GET /admin", h.GetAdmin)
	mux.HandleFunc("GET /admin/guest", h.GetAdminGuest)
	mux.HandleFunc("GET /posts/{id}/edit", h.GetPostEdit)
	mux.HandleFunc("GET /posts/{id}/preview", h.GetPostPreview)

	mux.HandleFunc("POST /posts", h.PostCreatePost)
	mux.HandleFunc("POST /posts/{id}/edit", h.PostEditPost)
	mux.HandleFunc("POST /posts/{id}/archive", h.PostArchivePost)
	mux.HandleFunc("POST /posts/{id}/unarchive", h.PostUnarchivePost)
	mux.HandleFunc("POST /posts/{id}/publish", h.PostPublishPost)
}

func (h *HTTP) GetHome(w http.ResponseWriter, r *http.Request) {
	posts, err := h.app.ListPublishedPosts(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	_ = views.Home(toPostCards(posts)).Render(r.Context(), w)
}

func (h *HTTP) GetPost(w http.ResponseWriter, r *http.Request) {
	post, err := h.app.GetPostBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	// drafts and archived posts only exist for the admin
	if post.Status != domain.PostStatusPublished {
		http.NotFound(w, r)
		return
	}

	_ = views.Post(toPostView(post)).Render(r.Context(), w)
}

func (h *HTTP) GetPostEdit(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	post, err := h.app.GetPostById(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	_ = views.Edit(post.ID.String(), string(post.MarkdownContent)).Render(r.Context(), w)
}

func (h *HTTP) GetPostPreview(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	post, err := h.app.GetPostById(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	_ = views.Preview(toPreviewView(post)).Render(r.Context(), w)
}

func (h *HTTP) GetAdmin(w http.ResponseWriter, r *http.Request) {
	posts, err := h.app.ListAllPosts(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	_ = views.Admin(toAdminPosts(posts)).Render(r.Context(), w)
}

func (h *HTTP) GetAdminGuest(w http.ResponseWriter, r *http.Request) {
	_ = views.GuestView().Render(r.Context(), w)
}

func (h *HTTP) PostCreatePost(w http.ResponseWriter, r *http.Request) {
	id, err := h.app.CreatePost(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	redirect(w, r, fmt.Sprintf("/posts/%s/edit", id.String()))
}

func (h *HTTP) PostPublishPost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	slug, err := h.app.PublishPost(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	redirect(w, r, fmt.Sprintf("/posts/%s", slug))
}

func (h *HTTP) PostArchivePost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = h.app.ArchivePost(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	redirect(w, r, "/admin")
}

func (h *HTTP) PostUnarchivePost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = h.app.UnarchivePost(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	redirect(w, r, "/admin")
}
func (h *HTTP) PostEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// TODO: read from body

	err = h.app.UpdateContent(r.Context(), id, nil)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	// TODO: update edit page
}

// redirect sends the client to url after a successful POST. htmx requests
// get HX-Redirect, since htmx would follow a 3xx inside the XHR and swap the
// target page in without changing the URL; plain form posts get a 303 so the
// browser follows up with a GET.
func redirect(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// writeError maps app errors onto status codes. Client errors carry their
// message, since it tells the author what to fix; anything else stays opaque
// to the client and gets logged instead.
func (h *HTTP) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, pkg.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, pkg.ErrConflict):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, pkg.ErrPreconditionNotMet),
		errors.Is(err, pkg.ErrEmpty),
		errors.Is(err, pkg.ErrExceedsMax),
		errors.Is(err, pkg.ErrBelowMin),
		errors.Is(err, pkg.ErrInvalidInput):
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	default:
		h.logger.ErrorContext(r.Context(), "internal error",
			"method", r.Method,
			"path", r.URL.Path,
			"err", err,
		)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// cacheImmutable marks everything under prefix as permanently cacheable
func cacheImmutable(h http.Handler, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, prefix) {
			w.Header().Set("Cache-Control", "max-age=31536000, immutable")
		}
		h.ServeHTTP(w, r)
	})
}
