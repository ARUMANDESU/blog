package app

import (
	"context"
	"time"
	"uuid"

	"github.com/arumandesu/blog/internal/domain"
	"github.com/arumandesu/blog/pkg"
)

type PostRepo interface {
	GetDomainPostById(context.Context, uuid.UUID) (*domain.Post, error)
	InsertPost(context.Context, *domain.Post) error
	UpdatePost(context.Context, *domain.Post) error

	GetPostById(context.Context, uuid.UUID) (Post, error)
	GetPostBySlug(context.Context, string) (Post, error)
	ListPosts(context.Context) ([]Post, error)
}

type App struct {
	postRepo  PostRepo
	txManager pkg.TxManager
}

type Post struct {
	ID          uuid.UUID
	Title       string
	Slug        string
	Description string
	HTMLContent []byte
	Status      domain.PostStatus
	CreatedAt   time.Time
}

func (a *App) CreatePost(ctx context.Context) (uuid.UUID, error) {
	post := domain.CreatePost()
	err := a.postRepo.InsertPost(ctx, post)
	if err != nil {
		return uuid.UUID{}, err
	}

	return post.Id(), nil
}

func (a *App) PublishPost(ctx context.Context, id uuid.UUID) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.postRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.Publish()
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

func (a *App) UpdateTitle(ctx context.Context, id uuid.UUID, title string) error {
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.postRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.UpdateTitle(title)
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
	return a.txManager.InTx(ctx, func(ctx context.Context) error {
		post, err := a.postRepo.GetDomainPostById(ctx, id)
		if err != nil {
			return err
		}

		err = post.UpdateContent(content)
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

		err = post.Unarchive()
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

func (a *App) GetPostById(ctx context.Context, id uuid.UUID) (Post, error) {
	return a.postRepo.GetPostById(ctx, id)
}
func (a *App) GetPostBySlug(ctx context.Context, slug string) (Post, error) {
	return a.postRepo.GetPostBySlug(ctx, slug)
}
func (a *App) ListPosts(ctx context.Context) ([]Post, error) {
	return a.postRepo.ListPosts(ctx)
}
