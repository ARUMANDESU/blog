package domain

import (
	"slices"
	"time"
	"uuid"

	"github.com/arumandesu/blog/pkg"
)

const (
	TitleMaxLen     = 200
	DescMaxLen      = 1500
	MdContentMaxLen = 100_000
)

type PostStatus string

const (
	PostStatusDraft    = PostStatus("draft")
	PostStatusPosted   = PostStatus("posted")
	PostStatusArchived = PostStatus("archived")
)

type Post struct {
	id              uuid.UUID
	title           string
	description     string
	markdownContent []byte
	htmlContent     []byte // fill this on write/update markdownContent
	attachedMedia   []uuid.UUID
	usedMedia       []uuid.UUID // all media attached to and imported into a post
	status          PostStatus
	createdAt       time.Time
	updatedAt       time.Time
	archivedAt      *time.Time
}

func NewID() uuid.UUID {
	return uuid.NewV7()
}

func CreatePost(
	id uuid.UUID,
	title string,
	description string,
	mdContent []byte,
	createdAt time.Time,
) (*Post, error) {
	if id == uuid.Nil() {
		return nil, pkg.NewFieldError("id", "must be provided", pkg.ErrEmpty)
	}
	if len(title) > TitleMaxLen {
		return nil, pkg.NewLimitError("title", TitleMaxLen, pkg.ErrExceedsMax)
	}
	if len(description) > DescMaxLen {
		return nil, pkg.NewLimitError("description", DescMaxLen, pkg.ErrExceedsMax)
	}
	// I am using len deliberately instead of [utf8.RuneCountInString]
	// because [ContentMaxLen] is just abuse guard not a real limit for user and len is O(1).
	// Expected content for this project is around 3k-30k chars:
	// which will fit en, ru and kz (note: chars of latter two takes 2 bytes in UTF-8)
	if len(mdContent) > MdContentMaxLen {
		return nil, pkg.NewLimitError("markdown_content", MdContentMaxLen, pkg.ErrExceedsMax)
	}
	if createdAt.IsZero() {
		return nil, pkg.NewFieldError("created_at", "must be provided", pkg.ErrEmpty)
	}
	createdAt = createdAt.UTC()

	return &Post{
		id:              id,
		title:           title,
		description:     description,
		markdownContent: mdContent,
		createdAt:       createdAt,
		updatedAt:       createdAt,
		archivedAt:      nil,
		attachedMedia:   nil,
		usedMedia:       nil,
		status:          PostStatusDraft,
	}, nil
}

func (p *Post) UpdateTitle(title string) error {
	if len(title) > TitleMaxLen {
		return pkg.NewLimitError("title", TitleMaxLen, pkg.ErrExceedsMax)
	}
	p.title = title
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

func (p *Post) UpdateContent(mdContent []byte, usedMedia []uuid.UUID) error {
	if len(mdContent) > MdContentMaxLen {
		return pkg.NewLimitError("markdown_content", MdContentMaxLen, pkg.ErrExceedsMax)
	}
	if slices.Contains(usedMedia, uuid.Nil()) {
		return pkg.NewFieldError("used_media", "all elements must be valid", pkg.ErrInvalidInput)
	}
	p.markdownContent = mdContent
	p.usedMedia = usedMedia
	// TODO: convert markdown into html and store it on update
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Post) UpdateAttachedMedia(attachedMedia []uuid.UUID, usedMedia []uuid.UUID) error {
	if slices.Contains(attachedMedia, uuid.Nil()) {
		return pkg.NewFieldError("attached_media", "all elements must be valid", pkg.ErrInvalidInput)
	}
	if slices.Contains(usedMedia, uuid.Nil()) {
		return pkg.NewFieldError("used_media", "all elements must be valid", pkg.ErrInvalidInput)
	}
	p.attachedMedia = attachedMedia
	p.usedMedia = usedMedia
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

type Media struct {
	id        uuid.UUID
	postId    uuid.UUID
	mime      string
	s3Key     string
	createdAt time.Time
	deletedAt *time.Time
}

func CreateMedia(
	id uuid.UUID,
	postId uuid.UUID,
	mime, s3Key string,
	createdAt time.Time,
) (*Media, error) {
	if id == uuid.Nil() {
		return nil, pkg.NewFieldError("id", "must be provided", pkg.ErrEmpty)
	}
	if len(mime) == 0 {
		return nil, pkg.NewFieldError("mime", "must be provided", pkg.ErrEmpty)
	}
	if len(s3Key) == 0 {
		return nil, pkg.NewFieldError("s3_key", "must be provided", pkg.ErrEmpty)
	}
	if createdAt.IsZero() {
		return nil, pkg.NewFieldError("created_at", "must be provided", pkg.ErrEmpty)
	}
	createdAt = createdAt.UTC()
	return &Media{
		id:        id,
		postId:    postId,
		mime:      mime,
		s3Key:     s3Key,
		createdAt: createdAt,
		deletedAt: nil,
	}, nil
}
