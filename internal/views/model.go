package views

import "time"

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

const dateLayout = "2006-01-02"

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
