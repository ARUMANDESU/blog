package app

import (
	"context"
	"io"
	"strings"
	"time"
	"uuid"

	"github.com/arumandesu/blog/internal/domain"
	"github.com/arumandesu/blog/pkg"
	slugx "github.com/gosimple/slug"
)

type PostRepo interface {
	GetDomainPostById(context.Context, uuid.UUID) (*domain.Post, error)
	InsertPost(context.Context, *domain.Post) error
	UpdatePost(context.Context, *domain.Post) error
	CheckSlug(context.Context, string) (bool, error)
}

type PostGetter interface {
	GetPostById(context.Context, uuid.UUID) (Post, error)
	GetPostBySlug(context.Context, string) (Post, error)
	ListPublishedPosts(context.Context) ([]Post, error)
	ListAllPosts(context.Context) ([]Post, error)
}

type MediaRepo interface {
	GetMediaById(context.Context, uuid.UUID) (*domain.Media, error)
	CreateMedia(context.Context, *domain.Media) error
}

type ObjectStorage interface {
	PutObject(ctx context.Context, r io.Reader, s3Key, mime string) error
}

type App struct {
	txManager     pkg.TxManager
	PostRepo      PostRepo
	PostGetter    PostGetter
	MediaRepo     MediaRepo
	ObjectStorage ObjectStorage
}

func New(txManager pkg.TxManager, postRepo PostRepo, postGetter PostGetter) *App {
	return &App{
		txManager:  txManager,
		PostRepo:   postRepo,
		PostGetter: postGetter,
	}
}

type Post struct {
	ID              uuid.UUID
	Title           string
	Slug            string
	Description     string
	MarkdownContent []byte
	HTMLContent     []byte
	Status          domain.PostStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (a *App) CreatePost(ctx context.Context) (uuid.UUID, error) {
	post := domain.CreatePost()
	err := a.PostRepo.InsertPost(ctx, post)
	if err != nil {
		return uuid.UUID{}, err
	}

	return post.Id(), nil
}

// PublishPost publishes a post and returns its slug
func (a *App) PublishPost(ctx context.Context, id uuid.UUID) (string, error) {
	var slug string
	err := a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.PostRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.Publish()
		if err != nil {
			return err
		}
		slug = post.Slug()

		err = a.PostRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
	return slug, err
}

func (a *App) UnpublishPost(ctx context.Context, id uuid.UUID) (string, error) {
	var slug string
	err := a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.PostRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		post.Unpublish()
		slug = post.Slug()

		err = a.PostRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
	return slug, err
}

// UpdateTitle updates title and slug, returns slug
func (a *App) UpdateTitle(ctx context.Context, id uuid.UUID, title string) (string, error) {
	var slug string
	err := a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.PostRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.UpdateTitle(title)
		if err != nil {
			return err
		}

		// TODO: update slug, if status is published then we should redirect old slug link into new one
		slug = truncateSlug(slugx.Make(title))
		if slug == "" {
			slug = post.Slug()
		} else {
			err = post.UpdateSlug(slug)
			if err != nil {
				return err
			}
		}

		err = a.PostRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
	return slug, err
}

const slugMaxLen = 45

// truncateSlug cuts slug to slugMaxLen and drops dashes left dangling by the cut
func truncateSlug(slug string) string {
	if len(slug) > slugMaxLen {
		slug = strings.Trim(slug[:slugMaxLen], "-")
	}
	return slug
}

func (a *App) UpdateSlug(ctx context.Context, id uuid.UUID, slug string) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.PostRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		if !slugx.IsSlug(slug) {
			return pkg.NewFieldError("slug", "invalid slug", domain.ErrInvalidSlug)
		}

		// TODO: update slug, if status is published then we should redirect old slug link into new one
		slug = truncateSlug(slug)
		err = post.UpdateSlug(slug)
		if err != nil {
			return err
		}

		err = a.PostRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *App) UpdateDescription(ctx context.Context, id uuid.UUID, description string) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.PostRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.UpdateDescription(description)
		if err != nil {
			return err
		}

		err = a.PostRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *App) UpdateContent(ctx context.Context, id uuid.UUID, content []byte) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.PostRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.UpdateContent(content)
		if err != nil {
			return err
		}

		err = a.PostRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *App) ArchivePost(ctx context.Context, id uuid.UUID) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.PostRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.Archive()
		if err != nil {
			return err
		}

		err = a.PostRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *App) UnarchivePost(ctx context.Context, id uuid.UUID) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.PostRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		post.Unarchive()

		err = a.PostRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
}

type UploadMediaDTO struct {
	R    io.Reader
	Ext  string
	Mime string
}

// UploadMedia uploads media to s3 and stores info in db, and returns s3Key
func (a *App) UploadMedia(ctx context.Context, dto UploadMediaDTO) (string, error) {
	s3key := uuid.New().String() + dto.Ext

	err := a.ObjectStorage.PutObject(ctx, dto.R, s3key, dto.Mime)
	if err != nil {
		return "", err
	}

	media, err := domain.CreateMedia(dto.Mime, s3key)
	if err != nil {
		return "", err
	}

	err = a.MediaRepo.CreateMedia(ctx, media)
	if err != nil {
		return "", err
	}

	return s3key, nil
}

func (a *App) GetPostById(ctx context.Context, id uuid.UUID) (Post, error) {
	return a.PostGetter.GetPostById(ctx, id)
}
func (a *App) GetPostBySlug(ctx context.Context, slug string) (Post, error) {
	return a.PostGetter.GetPostBySlug(ctx, slug)
}
func (a *App) ListPublishedPosts(ctx context.Context) ([]Post, error) {
	return a.PostGetter.ListPublishedPosts(ctx)
}
func (a *App) ListAllPosts(ctx context.Context) ([]Post, error) {
	return a.PostGetter.ListAllPosts(ctx)
}
