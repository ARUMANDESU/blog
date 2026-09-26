package transport

import (
	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/views"
)

func toPostCard(p app.Post) views.PostCard {
	return views.PostCard{
		Slug:        p.Slug,
		Title:       p.Title,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
	}
}

func toPostCards(ps []app.Post) []views.PostCard {
	cards := make([]views.PostCard, len(ps))
	for i, p := range ps {
		cards[i] = toPostCard(p)
	}
	return cards
}

func toPostView(p app.Post) views.PostView {
	return views.PostView{
		Title:       p.Title,
		CreatedAt:   p.CreatedAt,
		HTMLContent: string(p.HTMLContent),
	}
}

func toAdminPost(p app.Post) views.AdminPost {
	return views.AdminPost{
		ID:        p.ID.String(),
		Slug:      p.Slug,
		Title:     p.Title,
		Status:    string(p.Status),
		UpdatedAt: p.UpdatedAt,
	}
}

func toAdminPosts(ps []app.Post) []views.AdminPost {
	rows := make([]views.AdminPost, len(ps))
	for i, p := range ps {
		rows[i] = toAdminPost(p)
	}
	return rows
}

func toPreviewView(p app.Post) views.PreviewView {
	return views.PreviewView{
		ID:          p.ID.String(),
		Slug:        p.Slug,
		Title:       p.Title,
		Status:      string(p.Status),
		CreatedAt:   p.CreatedAt,
		HTMLContent: string(p.HTMLContent),
	}
}
