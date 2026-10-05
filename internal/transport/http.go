package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"uuid"

	"github.com/arumandesu/blog/assets"
	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/domain"
	"github.com/arumandesu/blog/internal/views"
	"github.com/arumandesu/blog/pkg"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const MaxBodySize = 5 << 20

var allowedMedia = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type HTTP struct {
	s3URL  string
	app    *app.App
	logger *slog.Logger
}

func NewHTTP(a *app.App, logger *slog.Logger, s3URL string) *HTTP {
	return &HTTP{app: a, logger: logger, s3URL: s3URL}
}

func Route(r chi.Router, h *HTTP) {
	fileServer := http.FileServerFS(assets.StaticFiles)

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(http.NewCrossOriginProtection().Handler)

	r.Get("/static/*", cacheImmutable(http.StripPrefix("/static", fileServer), "/static/fonts/"))

	r.Get("/", h.GetHome)

	r.Route("/posts", func(r chi.Router) {
		r.Get("/{slug}", h.GetPost)
		r.Get("/{id}/edit", h.GetPostEdit)
		r.Get("/{id}/preview", h.GetPostPreview)

		r.Post("/", h.PostCreatePost)
		r.Post("/{id}/archive", h.PostArchivePost)
		r.Post("/{id}/unarchive", h.PostUnarchivePost)
		r.Post("/{id}/publish", h.PostPublishPost)
		r.Post("/{id}/unpublish", h.PostUnpublishPost)

		r.Patch("/{id}/title", h.PatchPostTitle)
		r.Patch("/{id}/description", h.PatchPostDescription)
		r.Patch("/{id}/content", h.PatchPostContent)
		r.Patch("/{id}/slug", h.PatchPostSlug)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Get("/", h.GetAdmin)
		r.Get("/guest", h.GetAdminGuest)
	})

	r.Route("/media", func(r chi.Router) {
		r.Get("/{s3_key}", h.GetMedia)
		r.Post("/", h.PostUploadMedia)
	})
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

	_ = views.SlugInput(id.String(), slug, true).Render(r.Context(), w)
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

func (h *HTTP) GetMedia(w http.ResponseWriter, r *http.Request) {
	s3Key := r.PathValue("s3_key")
	s3Key = strings.TrimSpace(s3Key)
	if len(s3Key) == 0 {
		h.writeError(w, r, pkg.ErrInvalidInput)
		return
	}

	target := h.s3URL + "/" + url.PathEscape(s3Key)
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *HTTP) PostUploadMedia(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)
	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	contenType := http.DetectContentType(buf[:n])
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	ext, ok := allowedMedia[contenType]
	if !ok {
		http.Error(w, "unsupported file type", http.StatusUnsupportedMediaType)
		return
	}

	s3Key, err := h.app.UploadMedia(r.Context(), app.UploadMediaDTO{
		R:    file,
		Ext:  ext,
		Mime: contenType,
		Size: header.Size,
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	err = json.NewEncoder(w).Encode(map[string]string{"url": "/media/" + s3Key})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
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
func cacheImmutable(h http.Handler, prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, prefix) {
			w.Header().Set("Cache-Control", "max-age=31536000, immutable")
		}
		h.ServeHTTP(w, r)
	}
}
