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
	mux.HandleFunc("POST /posts/{id}/archive", h.PostArchivePost)
	mux.HandleFunc("POST /posts/{id}/unarchive", h.PostUnarchivePost)
	mux.HandleFunc("POST /posts/{id}/publish", h.PostPublishPost)
	mux.HandleFunc("POST /posts/{id}/unpublish", h.PostUnpublishPost)

	mux.HandleFunc("PATCH /posts/{id}/title", h.PatchPostTitle)
	mux.HandleFunc("PATCH /posts/{id}/description", h.PatchPostDescription)
	mux.HandleFunc("PATCH /posts/{id}/content", h.PatchPostContent)
	mux.HandleFunc("PATCH /posts/{id}/slug", h.PatchPostSlug)
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

	_ = views.Edit(toPostEdit(post)).Render(r.Context(), w)
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

func (h *HTTP) PostUnpublishPost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	_, err = h.app.UnpublishPost(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
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
func (h *HTTP) PatchPostContent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	content := r.FormValue("markdown-content")

	err = h.app.UpdateContent(r.Context(), id, []byte(content))
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HTTP) PatchPostTitle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	title := r.FormValue("title")
	title = strings.TrimSpace(title)

	slug, err := h.app.UpdateTitle(r.Context(), id, title)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	_ = views.Span("title-status", "Saved", "saved").Render(r.Context(), w)
	_ = views.SlugInput(slug, true).Render(r.Context(), w)
}

func (h *HTTP) PatchPostDescription(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	description := r.FormValue("description")
	description = strings.TrimSpace(description)

	err = h.app.UpdateDescription(r.Context(), id, description)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HTTP) PatchPostSlug(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	slug := r.FormValue("slug")
	slug = strings.TrimSpace(slug)

	err = h.app.UpdateSlug(r.Context(), id, slug)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
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
		errors.Is(err, pkg.ErrInvalidInput),
		errors.Is(err, domain.ErrInvalidSlug):
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
