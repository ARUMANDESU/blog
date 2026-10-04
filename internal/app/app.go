package app

import (
	"context"
	"io"
	"maps"
	"slices"
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
	BatchGetMediaByS3Key(ctx context.Context, s3Keys []string) ([]*domain.Media, error)
	BatchUpdateMedia(ctx context.Context, media []*domain.Media) error
	BatchDeleteMedia(ctx context.Context, ids []string) error
	GetMediaById(context.Context, uuid.UUID) (*domain.Media, error)
	GetUnusedMedia(context.Context) ([]*domain.Media, error)
	CreateMedia(context.Context, *domain.Media) error
}

type ObjectStorage interface {
	PutObject(ctx context.Context, r io.Reader, s3Key, mime string, size int64) error
	DeleteObjects(ctx context.Context, s3Keys []string) error
}

type App struct {
	txManager     pkg.TxManager
	postRepo      PostRepo
	postGetter    PostGetter
	mediaRepo     MediaRepo
	objectStorage ObjectStorage
}

func New(
	txManager pkg.TxManager,
	postRepo PostRepo,
	postGetter PostGetter,
	mediaRepo MediaRepo,
	objectStorage ObjectStorage,
) *App {
	return &App{
		txManager:     txManager,
		postRepo:      postRepo,
		postGetter:    postGetter,
		mediaRepo:     mediaRepo,
		objectStorage: objectStorage,
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
	err := a.postRepo.InsertPost(ctx, post)
	if err != nil {
		return uuid.UUID{}, err
	}

	return post.Id(), nil
}

// PublishPost publishes a post and returns its slug
func (a *App) PublishPost(ctx context.Context, id uuid.UUID) (string, error) {
	var slug string
	err := a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.postRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.Publish()
		if err != nil {
			return err
		}
		slug = post.Slug()

		err = a.postRepo.UpdatePost(ctx, post)
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
		post, err := a.postRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		post.Unpublish()
		slug = post.Slug()

		err = a.postRepo.UpdatePost(ctx, post)
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
		post, err := a.postRepo.GetDomainPostById(ctx, id)
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

		err = a.postRepo.UpdatePost(ctx, post)
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
		post, err := a.postRepo.GetDomainPostById(ctx, id)
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

		err = a.postRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *App) UpdateDescription(ctx context.Context, id uuid.UUID, description string) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.postRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.UpdateDescription(description)
		if err != nil {
			return err
		}

		err = a.postRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *App) UpdateContent(ctx context.Context, id uuid.UUID, content []byte) error {
	currRefsMap := pkg.MediaRefs(string(content))

	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.postRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		oldRefsMap := pkg.MediaRefs(string(post.MarkdownContent()))

		err = post.UpdateContent(content)
		if err != nil {
			return err
		}

		err = a.postRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}

		currMedia, err := a.fetchMedia(ctx, slices.Collect(maps.Keys(currRefsMap))) // current for self-healing
		if err != nil {
			return err
		}
		unusedMedia, err := a.fetchMedia(ctx, pkg.Difference(oldRefsMap, currRefsMap)) // old \ current
		if err != nil {
			return err
		}

		for _, m := range currMedia {
			m.Link(post.Id())
		}
		for _, m := range unusedMedia {
			m.Unlink()
		}

		return a.mediaRepo.BatchUpdateMedia(ctx, slices.Concat(currMedia, unusedMedia))
	})
}

func (a *App) fetchMedia(ctx context.Context, keys []string) ([]*domain.Media, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	return a.mediaRepo.BatchGetMediaByS3Key(ctx, keys)
}

func (a *App) ArchivePost(ctx context.Context, id uuid.UUID) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.postRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.Archive()
		if err != nil {
			return err
		}

		err = a.postRepo.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *App) UnarchivePost(ctx context.Context, id uuid.UUID) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.postRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		post.Unarchive()

		err = a.postRepo.UpdatePost(ctx, post)
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
	Size int64
}

// UploadMedia uploads media to s3 and stores info in db, and returns s3Key
func (a *App) UploadMedia(ctx context.Context, dto UploadMediaDTO) (string, error) {
	s3key := uuid.New().String() + dto.Ext

	err := a.objectStorage.PutObject(ctx, dto.R, s3key, dto.Mime, dto.Size)
	if err != nil {
		return "", err
	}

	media, err := domain.CreateMedia(dto.Mime, s3key)
	if err != nil {
		return "", err
	}

	err = a.mediaRepo.CreateMedia(ctx, media)
	if err != nil {
		return "", err
	}

	return s3key, nil
}

func (a *App) DeleteUnusedMedia(ctx context.Context) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		media, err := a.mediaRepo.GetUnusedMedia(ctx)
		if err != nil {
			return err
		}
		if len(media) == 0 {
			return nil
		}

		keys := make([]string, 0, len(media))
		ids := make([]string, 0, len(media))
		for _, m := range media {
			keys = append(keys, m.S3Key())
			ids = append(ids, m.Id().String())
		}

		err = a.objectStorage.DeleteObjects(ctx, keys)
		if err != nil {
			return err
		}

		return a.mediaRepo.BatchDeleteMedia(ctx, ids)
	})
}

func (a *App) GetPostById(ctx context.Context, id uuid.UUID) (Post, error) {
	return a.postGetter.GetPostById(ctx, id)
}
func (a *App) GetPostBySlug(ctx context.Context, slug string) (Post, error) {
	return a.postGetter.GetPostBySlug(ctx, slug)
}
func (a *App) ListPublishedPosts(ctx context.Context) ([]Post, error) {
	return a.postGetter.ListPublishedPosts(ctx)
}
func (a *App) ListAllPosts(ctx context.Context) ([]Post, error) {
	return a.postGetter.ListAllPosts(ctx)
}
