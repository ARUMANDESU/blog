package sqlite

import (
	"context"
	"database/sql"
	"time"
	"uuid"

	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/domain"
	"github.com/arumandesu/blog/internal/repository/sqlite/sqlcgen"
	"github.com/arumandesu/blog/pkg"
)

var _ app.MediaRepo = (*MediaRepo)(nil)

type MediaRepo struct {
	wdb, rdb *sql.DB
}

// BatchGetMediaByS3Key implements [app.MediaRepo].
func (m *MediaRepo) BatchGetMediaByS3Key(ctx context.Context, s3Keys []string) ([]*domain.Media, error) {
	if len(s3Keys) == 0 {
		return nil, nil
	}

	r := pkg.SqlConn(ctx, m.wdb)
	media, err := sqlcgen.New(r).GetMediaByS3Keys(ctx, s3Keys)
	if err != nil {
		return nil, pkg.WrapDBError(err)
	}

	return mediaToDomains(media)
}

// BatchUpdateMedia implements [app.MediaRepo].
func (m *MediaRepo) BatchUpdateMedia(ctx context.Context, media []*domain.Media) error {
	r := pkg.SqlConn(ctx, m.wdb)
	q := sqlcgen.New(r)

	for _, d := range media {
		m := domainToMedia(d)
		err := q.UpdateMedia(ctx, sqlcgen.UpdateMediaParams{
			PostID:    m.PostID,
			Mime:      m.Mime,
			S3Key:     m.S3Key,
			CreatedAt: m.CreatedAt,
			DeletedAt: m.DeletedAt,
			ID:        m.ID,
		})
		if err != nil {
			return pkg.WrapDBError(err)
		}
	}
	return nil
}

// BatchDeleteMedia implements [app.MediaRepo].
func (m *MediaRepo) BatchDeleteMedia(ctx context.Context, ids []string) error {
	r := pkg.SqlConn(ctx, m.wdb)
	err := sqlcgen.New(r).DeleteMediaByIds(ctx, ids)
	if err != nil {
		return pkg.WrapDBError(err)
	}
	return nil
}

// GetUnusedMedia implements [app.MediaRepo].
func (m *MediaRepo) GetUnusedMedia(ctx context.Context) ([]*domain.Media, error) {
	r := pkg.SqlConn(ctx, m.wdb)
	cutoff := formatTime(time.Now().Add(-time.Hour))
	media, err := sqlcgen.New(r).GetUnusedMedia(ctx, cutoff)
	if err != nil {
		return nil, pkg.WrapDBError(err)
	}
	return mediaToDomains(media)
}

// CreateMedia implements [app.MediaRepo].
func (m *MediaRepo) CreateMedia(ctx context.Context, d *domain.Media) error {
	r := pkg.SqlConn(ctx, m.wdb)
	media := domainToMedia(d)
	err := sqlcgen.New(r).CreateMedia(ctx, sqlcgen.CreateMediaParams(media))
	return pkg.WrapDBError(err)
}

// GetMediaById implements [app.MediaRepo].
func (m *MediaRepo) GetMediaById(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	r := pkg.SqlConn(ctx, m.wdb)
	media, err := sqlcgen.New(r).GetMedia(ctx, id.String())
	if err != nil {
		return nil, pkg.WrapDBError(err)
	}
	d, err := mediaToDomain(media)
	if err != nil {
		return nil, pkg.WrapDBError(err)
	}
	return d, nil
}

func NewMediaRepo(writeDB, readDB *sql.DB) *MediaRepo {
	return &MediaRepo{wdb: writeDB, rdb: readDB}
}
