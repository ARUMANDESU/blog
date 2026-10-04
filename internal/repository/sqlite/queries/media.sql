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

-- name: GetUnusedMedia :many
SELECT * FROM media
WHERE
    post_id IS NULL
    AND created_at < sqlc.arg('cutoff');

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

-- name: DeleteMediaByIds :exec
DELETE FROM media
WHERE id IN (sqlc.slice('ids'));
