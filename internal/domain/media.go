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
