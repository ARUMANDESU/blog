CREATE TABLE posts (
    id text PRIMARY KEY,
    title text NOT NULL,
    slug text NOT NULL,
    description text NOT NULL,
    markdown_content text,
    html_content text,
    status text NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
    created_at text NOT NULL,
    updated_at text NOT NULL
);

CREATE UNIQUE INDEX idx_uniq_posts_slug ON posts (slug);
