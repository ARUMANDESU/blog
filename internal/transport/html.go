package transport

import (
	"net/http"
	"strings"
	"time"

	"github.com/arumandesu/blog/assets"
	"github.com/arumandesu/blog/internal/views"
)

type HTTP struct {
}

func Handle(mux *http.ServeMux, h *HTTP) {
	fileServer := http.FileServerFS(assets.StaticFiles)

	mux.Handle("GET /static/", cacheImmutable(http.StripPrefix("/static", fileServer), "/static/fonts/"))
	mux.HandleFunc("GET /", h.GetHome)
	mux.HandleFunc("GET /posts/{slug}", h.GetPost)
	mux.HandleFunc("GET /posts/editor", h.GetPostsEditor)
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

func (h *HTTP) GetHome(w http.ResponseWriter, r *http.Request) {
	_ = views.Home(placeholderPosts).Render(r.Context(), w)
}

func (h *HTTP) GetPost(w http.ResponseWriter, r *http.Request) {
	_ = views.Post(placeholderPost).Render(r.Context(), w)
}

func (h *HTTP) GetPostsEditor(w http.ResponseWriter, r *http.Request) {
	_ = views.Editor().Render(r.Context(), w)
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
