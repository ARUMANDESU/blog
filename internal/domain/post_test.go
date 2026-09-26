package domain

import (
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/arumandesu/blog/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePost(t *testing.T) {
	t.Parallel()

	post := CreatePost()

	assert.NotEqual(t, uuid.UUID{}, post.id)
	assert.Empty(t, post.title)
	assert.Empty(t, post.slug)
	assert.Empty(t, post.description)
	assert.Nil(t, post.markdownContent)
	assert.Nil(t, post.htmlContent)
	assert.Equal(t, PostStatusDraft, post.status)
	assert.WithinDuration(t, time.Now().UTC(), post.createdAt, time.Second)
	assert.Equal(t, post.createdAt, post.updatedAt)
	assert.Nil(t, post.archivedAt)
}

func TestPost_UpdateTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		err   error
	}{
		{name: "valid title", title: "hello world"},
		{name: "empty title", title: ""},
		{name: "title at max length", title: strings.Repeat("a", TitleMaxLen)},
		{name: "title exceeds max length", title: strings.Repeat("a", TitleMaxLen+1), err: pkg.ErrExceedsMax},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			post := CreatePost()
			prevUpdatedAt := post.updatedAt

			err := post.UpdateTitle(tt.title)

			if tt.err != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
				assert.Empty(t, post.title)
				assert.Equal(t, prevUpdatedAt, post.updatedAt)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.title, post.title)
			assert.False(t, post.updatedAt.Before(prevUpdatedAt))
		})
	}
}

func TestPost_UpdateSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		slug string
		err  error
	}{
		{name: "valid slug", slug: "hello-world"},
		{name: "empty slug", slug: "", err: ErrInvalidSlug},
		{name: "slug with spaces and punctuation", slug: "Test & test", err: ErrInvalidSlug},
		{name: "slug starting with dash", slug: "-test", err: ErrInvalidSlug},
		{name: "slug ending with dash", slug: "test-", err: ErrInvalidSlug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			post := CreatePost()
			prevUpdatedAt := post.updatedAt

			err := post.UpdateSlug(tt.slug)

			if tt.err != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
				assert.Empty(t, post.slug)
				assert.Equal(t, prevUpdatedAt, post.updatedAt)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.slug, post.slug)
			assert.False(t, post.updatedAt.Before(prevUpdatedAt))
		})
	}
}

func TestPost_UpdateDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		desc string
		err  error
	}{
		{name: "valid description", desc: "test description"},
		{name: "empty description", desc: ""},
		{name: "description at max length", desc: strings.Repeat("a", DescMaxLen)},
		{name: "description exceeds max length", desc: strings.Repeat("a", DescMaxLen+1), err: pkg.ErrExceedsMax},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			post := CreatePost()
			prevUpdatedAt := post.updatedAt

			err := post.UpdateDescription(tt.desc)

			if tt.err != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
				assert.Empty(t, post.description)
				assert.Equal(t, prevUpdatedAt, post.updatedAt)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.desc, post.description)
			assert.False(t, post.updatedAt.Before(prevUpdatedAt))
		})
	}
}

func TestPost_UpdateContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mdContent []byte
		err       error
	}{
		{name: "valid content", mdContent: []byte("# Test Content")},
		{name: "empty content", mdContent: nil},
		{name: "content at max length", mdContent: []byte(strings.Repeat("a", MdContentMaxLen))},
		{name: "content exceeds max length", mdContent: []byte(strings.Repeat("a", MdContentMaxLen+1)), err: pkg.ErrExceedsMax},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			post := CreatePost()
			prevUpdatedAt := post.updatedAt

			err := post.UpdateContent(tt.mdContent)

			if tt.err != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
				assert.Nil(t, post.markdownContent)
				assert.Nil(t, post.htmlContent)
				assert.Equal(t, prevUpdatedAt, post.updatedAt)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.mdContent, post.markdownContent)
			assert.Equal(t, convertMd2HTML(tt.mdContent), post.htmlContent)
			assert.False(t, post.updatedAt.Before(prevUpdatedAt))
		})
	}
}

func TestPost_Archive(t *testing.T) {
	t.Parallel()

	post := CreatePost()
	prevUpdatedAt := post.updatedAt

	err := post.Archive()

	require.NoError(t, err)
	assert.Equal(t, PostStatusArchived, post.status)
	require.NotNil(t, post.archivedAt)
	assert.WithinDuration(t, time.Now().UTC(), *post.archivedAt, time.Second)
	assert.Equal(t, *post.archivedAt, post.updatedAt)
	assert.False(t, post.updatedAt.Before(prevUpdatedAt))
}

func TestPost_Unarchive(t *testing.T) {
	t.Parallel()

	post := CreatePost()
	require.NoError(t, post.Archive())
	prevUpdatedAt := post.updatedAt

	err := post.Unarchive()

	require.NoError(t, err)
	assert.Equal(t, PostStatusDraft, post.status)
	assert.Nil(t, post.archivedAt)
	assert.False(t, post.updatedAt.Before(prevUpdatedAt))
}

func TestPost_Publish(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		desc  string
		slug  string
		err   error
	}{
		{name: "all fields set", title: "test", desc: "test description", slug: "test"},
		{name: "missing title", desc: "test description", slug: "test", err: pkg.ErrPreconditionNotMet},
		{name: "missing description", title: "test", slug: "test", err: pkg.ErrPreconditionNotMet},
		{name: "missing title and description", slug: "test", err: pkg.ErrPreconditionNotMet},
		{name: "missing slug", title: "test", desc: "test description", err: ErrEmptySlug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			post := CreatePost()
			require.NoError(t, post.UpdateTitle(tt.title))
			require.NoError(t, post.UpdateDescription(tt.desc))
			if tt.slug != "" {
				require.NoError(t, post.UpdateSlug(tt.slug))
			}
			prevUpdatedAt := post.updatedAt

			err := post.Publish()

			if tt.err != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
				assert.Equal(t, PostStatusDraft, post.status)
				assert.Equal(t, prevUpdatedAt, post.updatedAt)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, PostStatusPublished, post.status)
			assert.False(t, post.updatedAt.Before(prevUpdatedAt))
		})
	}
}

func TestPost_Draft(t *testing.T) {
	t.Parallel()

	post := CreatePost()
	require.NoError(t, post.UpdateTitle("test"))
	require.NoError(t, post.UpdateDescription("test description"))
	require.NoError(t, post.UpdateSlug("test"))
	require.NoError(t, post.Publish())
	prevUpdatedAt := post.updatedAt

	err := post.Draft()

	require.NoError(t, err)
	assert.Equal(t, PostStatusDraft, post.status)
	assert.False(t, post.updatedAt.Before(prevUpdatedAt))
}

func TestPost_Id(t *testing.T) {
	t.Parallel()

	post := CreatePost()

	assert.Equal(t, post.id, post.Id())
}
