package transport

import (
	"net/http"
	"strings"
	"time"
	"uuid"

	"github.com/arumandesu/blog/assets"
	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/views"
)

type HTTP struct {
	app *app.App
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
	mux.HandleFunc("POST /posts/{id}/post", h.PostPostPost)
}

// TODO: replace the placeholder data

var placeholderPosts = []views.PostCard{
	{
		Slug:        "hello",
		Title:       "Writing a blog engine in Go",
		Description: "templ for the templates, htmx for the interactions, no build step and no JavaScript framework anywhere in sight.",
		CreatedAt:   time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
	},
	{
		Slug:        "domain",
		Title:       "Keeping the domain honest",
		Description: "Unexported fields, constructors that validate, and why the HTTP layer never gets to reach inside an aggregate.",
		CreatedAt:   time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
	},
	{
		Slug:        "embed",
		Title:       "Shipping one binary",
		Description: "go:embed for the static assets, so deploying the site is a single file copy.",
		CreatedAt:   time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
	},
}

var placeholderPost = views.PostView{
	Title:     "Writing a blog engine in Go",
	CreatedAt: time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
	HTMLContent: `<p>This is placeholder content standing in for rendered markdown.</p>
<h2>Why templ</h2>
<p>Templates are compiled Go, so a typo is a build error instead of a blank page at runtime.</p>
<pre><code>templ Home(posts []PostCard) {
	@Base("ARUMANDESU") {
		&lt;ul class="post-list"&gt;...&lt;/ul&gt;
	}
}</code></pre>
<blockquote><p>The nice part: no separate template cache, no reflection.</p></blockquote>
<h2>What is next</h2>
<ul><li>Markdown to HTML on write</li><li>Media uploads</li><li>An admin surface</li></ul>`,
}

// placeholderMarkdown is the source placeholderPost.HTMLContent stands in for.
const placeholderMarkdown = "This is placeholder content standing in for rendered markdown.\n" +
	"\n" +
	"## Why templ\n" +
	"\n" +
	"Templates are compiled Go, so a typo is a build error instead of a blank page at runtime.\n" +
	"\n" +
	"```templ\n" +
	"templ Home(posts []PostCard) {\n" +
	"\t@Base(\"ARUMANDESU\") {\n" +
	"\t\t<ul class=\"post-list\">...</ul>\n" +
	"\t}\n" +
	"}\n" +
	"```\n" +
	"\n" +
	"> The nice part: no separate template cache, no reflection.\n" +
	"\n" +
	"## What is next\n" +
	"\n" +
	"- Markdown to HTML on write\n" +
	"- Media uploads\n" +
	"- An admin surface\n"

var placeholderAdminPosts = []views.AdminPost{
	{ID: "0199a1b2-0000-7000-8000-000000000003", Title: "", Status: views.StatusDraft, UpdatedAt: time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC)},
	{ID: "0199a1b2-0000-7000-8000-000000000001", Slug: "hello", Title: "Writing a blog engine in Go", Status: views.StatusPosted, UpdatedAt: time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC)},
	{ID: "0199a1b2-0000-7000-8000-000000000002", Slug: "domain", Title: "Keeping the domain honest", Status: views.StatusArchived, UpdatedAt: time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)},
}

var placeholderPreview = views.PreviewView{
	Title:       placeholderPost.Title,
	Status:      views.StatusDraft,
	CreatedAt:   placeholderPost.CreatedAt,
	HTMLContent: placeholderPost.HTMLContent,
}

func (h *HTTP) GetHome(w http.ResponseWriter, r *http.Request) {
	_ = views.Home(placeholderPosts).Render(r.Context(), w)
}

func (h *HTTP) GetPost(w http.ResponseWriter, r *http.Request) {
	_ = views.Post(placeholderPost).Render(r.Context(), w)
}

func (h *HTTP) GetPostEdit(w http.ResponseWriter, r *http.Request) {
	_ = views.Edit(r.PathValue("id"), placeholderMarkdown).Render(r.Context(), w)
}

func (h *HTTP) GetPostPreview(w http.ResponseWriter, r *http.Request) {
	p := placeholderPreview
	p.ID = r.PathValue("id")
	_ = views.Preview(p).Render(r.Context(), w)
}

func (h *HTTP) GetAdmin(w http.ResponseWriter, r *http.Request) {
	_ = views.Admin(placeholderAdminPosts).Render(r.Context(), w)
}

func (h *HTTP) GetAdminGuest(w http.ResponseWriter, r *http.Request) {
	_ = views.GuestView().Render(r.Context(), w)
}

func (h *HTTP) PostCreatePost(w http.ResponseWriter, r *http.Request) {
	_, err := h.app.CreatePost(r.Context())
	if err != nil {
		// TODO: handle error
		return
	}

	// TODO: redirect to /posts/{id}/edit
}

func (h *HTTP) PostPostPost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		// TODO: handle error
		return
	}

	err = h.app.PostPost(r.Context(), id)
	if err != nil {
		// TODO: handle error
		return
	}

	// TODO: redirect to /posts/{id}
}

func (h *HTTP) PostArchivePost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		// TODO: handle error
		return
	}

	err = h.app.ArchivePost(r.Context(), id)
	if err != nil {
		// TODO: handle error
		return
	}

	// TODO: redirect to /admin
}
func (h *HTTP) PostEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		// TODO: handle error
		return
	}
	// TODO: read from body

	err = h.app.UpdateContent(r.Context(), id, nil)
	if err != nil {
		// TODO: handle error
		return
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
