package main

import (
	"context"
	"fmt"

	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/domain"
)

type seedPost struct {
	title       string
	slug        string
	description string
	markdown    string
	status      domain.PostStatus
}

var seedPosts = []seedPost{
	{
		title:       "Shipping one binary",
		slug:        "embed",
		description: "go:embed for the static assets, so deploying the site is a single file copy.",
		markdown:    "Everything under `assets/` is embedded with `go:embed`, so the deploy is one file.\n",
		status:      domain.PostStatusPublished,
	},
	{
		title:       "An old post",
		slug:        "old",
		description: "Something that no longer holds up.",
		markdown:    "Archived, so only the admin can see it.\n",
		status:      domain.PostStatusArchived,
	},
	{
		title:       "Keeping the domain honest",
		slug:        "domain",
		description: "Unexported fields, constructors that validate, and why the HTTP layer never gets to reach inside an aggregate.",
		markdown:    "The aggregate owns its invariants; everything outside talks to it through methods.\n",
		status:      domain.PostStatusPublished,
	},
	{
		title:       "Writing a blog engine in Go",
		slug:        "hello",
		description: "templ for the templates, htmx for the interactions, no build step and no JavaScript framework anywhere in sight.",
		markdown: "This is placeholder content standing in for rendered markdown.\n" +
			"\n" +
			"## Why templ\n" +
			"\n" +
			"Templates are compiled Go, so a typo is a build error instead of a blank page at runtime.\n" +
			"\n" +
			"```templ\n" +
			"templ Home(posts []PostCard) {\n" +
			"\t@Base(\"ARUMANDESU\") {\n" +
			"\t\t<ul class=\"post-list\">...</ul>\n" +
			"\t}\n" +
			"}\n" +
			"```\n" +
			"\n" +
			"> The nice part: no separate template cache, no reflection.\n" +
			"\n" +
			"## What is next\n" +
			"\n" +
			"- Markdown to HTML on write\n" +
			"- Media uploads\n" +
			"- An admin surface\n",
		status: domain.PostStatusPublished,
	},
	{
		title:       "Media uploads",
		slug:        "media",
		description: "Work in progress.",
		markdown:    "Not ready yet.\n",
		status:      domain.PostStatusDraft,
	},
	// freshly created, still empty draft
	{status: domain.PostStatusDraft},
}

func seed(ctx context.Context, repo app.PostRepo) error {
	for _, s := range seedPosts {
		post, err := s.build()
		if err != nil {
			return fmt.Errorf("seed %q: %w", s.slug, err)
		}
		if err := repo.InsertPost(ctx, post); err != nil {
			return fmt.Errorf("seed %q: %w", s.slug, err)
		}
	}
	return nil
}

func (s seedPost) build() (*domain.Post, error) {
	post := domain.CreatePost()
	if s.title == "" {
		return post, nil
	}

	if err := post.UpdateTitle(s.title); err != nil {
		return nil, err
	}
	if err := post.UpdateSlug(s.slug); err != nil {
		return nil, err
	}
	if err := post.UpdateDescription(s.description); err != nil {
		return nil, err
	}
	if err := post.UpdateContent([]byte(s.markdown)); err != nil {
		return nil, err
	}

	switch s.status {
	case domain.PostStatusPublished:
		err := post.Publish()
		return post, err
	case domain.PostStatusArchived:
		if err := post.Publish(); err != nil {
			return nil, err
		}
		err := post.Archive()
		return post, err
	}
	return post, nil
}
