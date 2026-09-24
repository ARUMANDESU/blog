package views

import (
	"time"

	"github.com/a-h/templ"
)

// PostCard is what the home page needs to render one entry in the list.
type PostCard struct {
	Slug        string
	Title       string
	Description string
	CreatedAt   time.Time
}

// PostView is a single rendered post. HTMLContent is the already-converted
// markdown, injected raw — sanitize it before it reaches here.
type PostView struct {
	Title       string
	CreatedAt   time.Time
	HTMLContent string
}

// Post statuses as the admin views see them; mirrors domain.PostStatus.
const (
	StatusDraft    = "draft"
	StatusPosted   = "posted"
	StatusArchived = "archived"
)

// AdminPost is one row in the admin post list.
type AdminPost struct {
	ID        string
	Slug      string
	Title     string
	Status    string
	UpdatedAt time.Time
}

// PreviewView is a post rendered as readers would see it, plus what the
// admin actions next to it need.
type PreviewView struct {
	ID          string
	Slug        string
	Title       string
	Status      string
	CreatedAt   time.Time
	HTMLContent string
}

const dateLayout = "2006-01-02"

func postEditURL(id string) templ.SafeURL {
	return templ.SafeURL("/posts/" + id + "/edit")
}

func postPreviewURL(id string) templ.SafeURL {
	return templ.SafeURL("/posts/" + id + "/preview")
}

func postPublishURL(id string) templ.SafeURL {
	return templ.SafeURL("/posts/" + id + "/publish")
}

func postArchiveURL(id string) templ.SafeURL {
	return templ.SafeURL("/posts/" + id + "/archive")
}

// titleOrUntitled keeps freshly created, still empty drafts clickable.
func titleOrUntitled(title string) string {
	if title == "" {
		return "(untitled)"
	}
	return title
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(dateLayout)
}

// machineDate feeds <time datetime="...">, so crawlers and readers agree.
func machineDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
