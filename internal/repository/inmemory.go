package repository

import (
	"cmp"
	"context"
	"slices"
	"sync"
	"uuid"

	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/domain"
	"github.com/arumandesu/blog/pkg"
)

var _ app.PostRepo = (*InMemoryPostRepo)(nil)

type InMemoryPostRepo struct {
	mu    sync.RWMutex
	posts map[uuid.UUID]domain.Post
}

func NewInMemoryPostRepo() *InMemoryPostRepo {
	return &InMemoryPostRepo{posts: make(map[uuid.UUID]domain.Post)}
}

func (r *InMemoryPostRepo) GetDomainPostById(_ context.Context, id uuid.UUID) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, ok := r.posts[id]
	if !ok {
		return nil, pkg.ErrNotFound
	}
	return &post, nil
}

func (r *InMemoryPostRepo) InsertPost(_ context.Context, post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.posts[post.Id()]; ok {
		return pkg.ErrConflict
	}
	if err := r.checkSlugUnique(post); err != nil {
		return err
	}
	r.posts[post.Id()] = *post
	return nil
}

func (r *InMemoryPostRepo) UpdatePost(_ context.Context, post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.posts[post.Id()]; !ok {
		return pkg.ErrNotFound
	}
	if err := r.checkSlugUnique(post); err != nil {
		return err
	}
	r.posts[post.Id()] = *post
	return nil
}

func (r *InMemoryPostRepo) GetPostById(_ context.Context, id uuid.UUID) (app.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, ok := r.posts[id]
	if !ok {
		return app.Post{}, pkg.ErrNotFound
	}
	return toAppPost(&post), nil
}

func (r *InMemoryPostRepo) GetPostBySlug(_ context.Context, slug string) (app.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if slug == "" {
		return app.Post{}, pkg.ErrNotFound
	}
	for _, post := range r.posts {
		if post.Slug() == slug {
			return toAppPost(&post), nil
		}
	}
	return app.Post{}, pkg.ErrNotFound
}

func (r *InMemoryPostRepo) ListPosts(_ context.Context) ([]app.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	posts := make([]app.Post, 0, len(r.posts))
	for _, post := range r.posts {
		posts = append(posts, toAppPost(&post))
	}
	// newest first
	slices.SortFunc(posts, func(a, b app.Post) int {
		return cmp.Compare(b.CreatedAt.UnixNano(), a.CreatedAt.UnixNano())
	})
	return posts, nil
}

// checkSlugUnique must be called with r.mu held.
func (r *InMemoryPostRepo) checkSlugUnique(post *domain.Post) error {
	if post.Slug() == "" {
		return nil
	}
	for id, other := range r.posts {
		if id != post.Id() && other.Slug() == post.Slug() {
			return pkg.ErrConflict
		}
	}
	return nil
}

func toAppPost(p *domain.Post) app.Post {
	return app.Post{
		ID:          p.Id(),
		Title:       p.Title(),
		Slug:        p.Slug(),
		Description: p.Description(),
		HTMLContent: p.HTMLContent(),
		Status:      p.Status(),
		CreatedAt:   p.CreatedAt(),
	}
}
