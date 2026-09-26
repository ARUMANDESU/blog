package domain

import (
	"bytes"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/arumandesu/blog/pkg"
	slugx "github.com/gosimple/slug"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

const (
	TitleMaxLen     = 200
	DescMaxLen      = 1500
	MdContentMaxLen = 100_000
)

var (
	ErrInvalidSlug = errors.New("invalid slug")
	ErrEmptySlug   = errors.New("empty slug")
)

type PostStatus string

const (
	PostStatusDraft     = PostStatus("draft")
	PostStatusPublished = PostStatus("posted")
	PostStatusArchived  = PostStatus("archived")
)

type Post struct {
	id              uuid.UUID
	title           string
	slug            string
	description     string
	markdownContent []byte
	htmlContent     []byte // fill this on write/update markdownContent
	status          PostStatus
	createdAt       time.Time
	updatedAt       time.Time
	archivedAt      *time.Time
}

func NewID() uuid.UUID {
	return uuid.NewV7()
}

func CreatePost() *Post {
	now := time.Now().UTC()
	return &Post{
		id:         NewID(),
		status:     PostStatusDraft,
		createdAt:  now,
		updatedAt:  now,
		archivedAt: nil,
	}
}

func (p *Post) UpdateTitle(title string) error {
	if len(title) > TitleMaxLen {
		return pkg.NewLimitError("title", TitleMaxLen, pkg.ErrExceedsMax)
	}
	p.title = title
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Post) UpdateSlug(slug string) error {
	if !slugx.IsSlug(slug) {
		return ErrInvalidSlug
	}
	p.slug = slug
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Post) UpdateDescription(desc string) error {
	if len(desc) > DescMaxLen {
		return pkg.NewLimitError("description", DescMaxLen, pkg.ErrExceedsMax)
	}
	p.description = desc
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Post) UpdateContent(mdContent []byte) error {
	if len(mdContent) > MdContentMaxLen {
		return pkg.NewLimitError("markdown_content", MdContentMaxLen, pkg.ErrExceedsMax)
	}
	p.markdownContent = mdContent
	p.htmlContent = convertMd2HTML(mdContent)
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Post) Archive() error {
	now := time.Now().UTC()
	p.updatedAt = now
	p.archivedAt = &now
	p.status = PostStatusArchived
	return nil
}
func (p *Post) Unarchive() error {
	p.archivedAt = nil
	return p.Draft()
}

func (p *Post) Publish() error {
	var err error
	if len(p.title) == 0 {
		err = fmt.Errorf("%w: title must be provided", pkg.ErrPreconditionNotMet)
	}
	if len(p.description) == 0 {
		err = errors.Join(err, fmt.Errorf("%w: description must be provided", pkg.ErrPreconditionNotMet))
	}
	if err != nil {
		return err
	}
	if len(p.slug) == 0 {
		return ErrEmptySlug
	}

	p.updatedAt = time.Now().UTC()
	p.status = PostStatusPublished
	return nil
}

func (p *Post) Draft() error {
	p.updatedAt = time.Now().UTC()
	p.status = PostStatusDraft
	return nil
}

func (p *Post) Id() uuid.UUID        { return p.id }
func (p *Post) Title() string        { return p.title }
func (p *Post) Slug() string         { return p.slug }
func (p *Post) Description() string  { return p.description }
func (p *Post) HTMLContent() []byte  { return p.htmlContent }
func (p *Post) Status() PostStatus   { return p.status }
func (p *Post) CreatedAt() time.Time { return p.createdAt }

func convertMd2HTML(mdContent []byte) []byte {
	var buf bytes.Buffer
	p := parser.New()
	n := p.Parse(mdContent)
	r := html.New()
	// I am intentionally ignoring this error because
	// we are using [bytes.Buffer], and we are
	// technically safe to ignore
	_ = r.Render(&buf, mdContent, n)
	return buf.Bytes()
}
