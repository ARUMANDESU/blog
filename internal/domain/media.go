package domain

import (
	"time"
	"uuid"

	"github.com/arumandesu/blog/pkg"
)

type Media struct {
	id        uuid.UUID
	postId    uuid.UUID
	mime      string
	s3Key     string
	createdAt time.Time
	deletedAt *time.Time
}

func CreateMedia(mime, s3Key string) (*Media, error) {
	if len(mime) == 0 {
		return nil, pkg.NewFieldError("mime", "must be provided", pkg.ErrEmpty)
	}
	if len(s3Key) == 0 {
		return nil, pkg.NewFieldError("s3_key", "must be provided", pkg.ErrEmpty)
	}
	return &Media{
		id:        NewID(),
		postId:    uuid.Nil(),
		mime:      mime,
		s3Key:     s3Key,
		createdAt: time.Now().UTC(),
		deletedAt: nil,
	}, nil
}

func (m *Media) Link(postId uuid.UUID) { m.postId = postId }
func (m *Media) Unlink()               { m.postId = uuid.Nil() }

func (m *Media) Id() uuid.UUID         { return m.id }
func (m *Media) PostId() uuid.UUID     { return m.postId }
func (m *Media) MIME() string          { return m.mime }
func (m *Media) S3Key() string         { return m.s3Key }
func (m *Media) CreatedAt() time.Time  { return m.createdAt }
func (m *Media) DeletedAt() *time.Time { return m.deletedAt }

type UnmarshalMediaDBDTO struct {
	ID        uuid.UUID
	PostId    uuid.UUID
	Mime      string
	S3Key     string
	CreatedAt time.Time
	DeletedAt *time.Time
}

// UnmarshalMediaDB convert dto into media
//
// WARNING: this can be used only by repository layer
func UnmarshalMediaDB(dto UnmarshalMediaDBDTO) *Media {
	return &Media{
		id:        dto.ID,
		postId:    dto.PostId,
		mime:      dto.Mime,
		s3Key:     dto.S3Key,
		createdAt: dto.CreatedAt,
		deletedAt: dto.DeletedAt,
	}
}
