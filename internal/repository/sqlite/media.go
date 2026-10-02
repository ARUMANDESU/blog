package sqlite

import (
	"context"
	"database/sql"
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
