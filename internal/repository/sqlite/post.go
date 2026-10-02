package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"uuid"

	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/domain"
	"github.com/arumandesu/blog/internal/repository/sqlite/sqlcgen"
	"github.com/arumandesu/blog/pkg"
)

var _ app.PostGetter = (*PostRepo)(nil)
var _ app.PostRepo = (*PostRepo)(nil)

type PostRepo struct {
	wdb *sql.DB
	rdb *sql.DB
}

func NewPostRepo(writeDB, readDB *sql.DB) *PostRepo {
	return &PostRepo{
		wdb: writeDB,
		rdb: readDB,
	}
}

// CheckSlug implements [app.PostRepo].
func (p *PostRepo) CheckSlug(ctx context.Context, slug string) (bool, error) {
	r := pkg.SqlConn(ctx, p.wdb)
	_, err := sqlcgen.New(r).GetPostBySlug(ctx, stringToNull(slug))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, pkg.WrapDBError(err)
	}

	return true, nil
}

// GetDomainPostById implements [app.PostRepo].
func (p *PostRepo) GetDomainPostById(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	r := pkg.SqlConn(ctx, p.wdb)

	post, err := sqlcgen.New(r).GetPost(ctx, id.String())
	if err != nil {
		return nil, pkg.WrapDBError(err)
	}

	return postToDomain(post)
}

// InsertPost implements [app.PostRepo].
func (p *PostRepo) InsertPost(ctx context.Context, d *domain.Post) error {
	r := pkg.SqlConn(ctx, p.wdb)
	post := domainToPost(d)
	err := sqlcgen.New(r).CreatePost(ctx, sqlcgen.CreatePostParams(post))
	return pkg.WrapDBError(err)
}

// UpdatePost implements [app.PostRepo].
func (p *PostRepo) UpdatePost(ctx context.Context, d *domain.Post) error {
	r := pkg.SqlConn(ctx, p.wdb)
	post := domainToPost(d)
	err := sqlcgen.New(r).UpdatePost(ctx, sqlcgen.UpdatePostParams{
		ID:              post.ID,
		Title:           post.Title,
		Slug:            post.Slug,
		Description:     post.Description,
		MarkdownContent: post.MarkdownContent,
		HtmlContent:     post.HtmlContent,
		Status:          post.Status,
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
	})
	return pkg.WrapDBError(err)
}

// GetPostById implements [app.PostGetter].
func (p *PostRepo) GetPostById(ctx context.Context, id uuid.UUID) (app.Post, error) {
	post, err := sqlcgen.New(p.rdb).GetPost(ctx, id.String())
	if err != nil {
		return app.Post{}, pkg.WrapDBError(err)
	}

	return postToApp(post)
}

// GetPostBySlug implements [app.PostGetter].
func (p *PostRepo) GetPostBySlug(ctx context.Context, slug string) (app.Post, error) {
	post, err := sqlcgen.New(p.rdb).GetPostBySlug(ctx, stringToNull(slug))
	if err != nil {
		return app.Post{}, pkg.WrapDBError(err)
	}

	return postToApp(post)
}

// ListAllPosts implements [app.PostGetter].
func (p *PostRepo) ListAllPosts(ctx context.Context) ([]app.Post, error) {
	posts, err := sqlcgen.New(p.rdb).ListAllPosts(ctx)
	if err != nil {
		return nil, pkg.WrapDBError(err)
	}

	return postsToApp(posts)
}

// ListPublishedPosts implements [app.PostGetter].
func (p *PostRepo) ListPublishedPosts(ctx context.Context) ([]app.Post, error) {
	posts, err := sqlcgen.New(p.rdb).ListPublishedPosts(ctx)
	if err != nil {
		return nil, pkg.WrapDBError(err)
	}

	return postsToApp(posts)
}
