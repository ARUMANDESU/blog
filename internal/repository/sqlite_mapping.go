package repository

import (
	"database/sql"
	"time"
	"uuid"

	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/domain"
	"github.com/arumandesu/blog/internal/repository/sqlite/sqlcgen"
)

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func bytesToSqlNullString(b []byte) sql.NullString {
	var ns sql.NullString
	if b != nil {
		ns.Valid = true
		ns.String = string(b)
	}
	return ns
}

func NullStringToBytes(ns sql.NullString) []byte {
	var b []byte
	if ns.Valid {
		b = []byte(ns.String)
	}
	return b
}

func stringToNull(s string) sql.NullString {
	var ns sql.NullString
	if len(s) > 0 {
		ns.Valid = true
		ns.String = s
	}
	return ns
}

func domainToPost(d *domain.Post) sqlcgen.Post {
	return sqlcgen.Post{
		ID:              d.Id().String(),
		Title:           d.Title(),
		Slug:            stringToNull(d.Slug()),
		Description:     d.Description(),
		MarkdownContent: bytesToSqlNullString(d.MarkdownContent()),
		HtmlContent:     bytesToSqlNullString(d.HTMLContent()),
		Status:          string(d.Status()),
		CreatedAt:       formatTime(d.CreatedAt()),
		UpdatedAt:       formatTime(d.UpdatedAt()),
	}
}

func postToDomain(p sqlcgen.Post) (*domain.Post, error) {
	createdAt, err := parseTime(p.CreatedAt)
	if err != nil {
		return nil, err
	}

	updatedAt, err := parseTime(p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return domain.UnmarshalDB(domain.UnmarshalDBDTO{
		ID:              uuid.MustParse(p.ID),
		Title:           p.Title,
		Slug:            p.Slug.String, // if not valid then zero string
		Description:     p.Description,
		MarkdownContent: NullStringToBytes(p.MarkdownContent),
		HTMLContent:     NullStringToBytes(p.HtmlContent),
		Status:          domain.PostStatus(p.Status),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}), nil
}

func postToApp(p sqlcgen.Post) (app.Post, error) {
	var mdContent, htmlContent []byte
	if p.MarkdownContent.Valid {
		mdContent = []byte(p.MarkdownContent.String)
	}
	if p.HtmlContent.Valid {
		htmlContent = []byte(p.HtmlContent.String)
	}
	createdAt, err := parseTime(p.CreatedAt)
	if err != nil {
		return app.Post{}, err
	}

	updatedAt, err := parseTime(p.UpdatedAt)
	if err != nil {
		return app.Post{}, err
	}

	return app.Post{
		ID:              uuid.MustParse(p.ID),
		Title:           p.Title,
		Slug:            p.Slug.String, // if not valid then zero string
		Description:     p.Description,
		MarkdownContent: mdContent,
		HTMLContent:     htmlContent,
		Status:          domain.PostStatus(p.Status),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}

func postsToApp(ps []sqlcgen.Post) ([]app.Post, error) {
	as := make([]app.Post, 0, len(ps))

	for _, p := range ps {
		a, err := postToApp(p)
		if err != nil {
			return nil, err
		}

		as = append(as, a)
	}

	return as, nil
}
