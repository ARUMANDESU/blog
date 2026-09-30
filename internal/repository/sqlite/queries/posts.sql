-- name: GetPost :one
SELECT *
FROM posts
WHERE id = ?;

-- name: GetPostBySlug :one
SELECT *
FROM posts
WHERE slug = ?;

-- name: ListAllPosts :many
SELECT *
FROM posts
ORDER BY created_at DESC, id DESC;

-- name: ListPublishedPosts :many
SELECT *
FROM posts
WHERE status = 'published';

-- name: CreatePost :exec
INSERT INTO posts (
    id, title, slug, description,
    markdown_content, html_content,
    status, created_at, updated_at
) VALUES (?,?,?,?,?,?,?,?,?);

-- name: UpdatePost :exec
UPDATE posts 
SET title = ?, slug = ?, description = ?,
markdown_content = ?, html_content = ?,
status = ?, created_at = ?, updated_at = ?
WHERE id = ?;
