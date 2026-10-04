-- name: GetMedia :one
SELECT *
FROM media
WHERE id = ?;

-- name: GetMediaByS3Key :one
SELECT *
FROM media
WHERE s3_key = ?;

-- name: GetMediaByS3Keys :many
SELECT * FROM media
WHERE s3_key IN (sqlc.slice('keys'));

-- name: CreateMedia :exec
INSERT INTO media (
    id, post_id, mime, s3_key,
    created_at, deleted_at
) VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateMedia :exec
UPDATE media
SET post_id = ?, mime = ?, s3_key = ?,
created_at = ?, deleted_at = ?
WHERE id = ?;
