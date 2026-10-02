CREATE TABLE media (
    id text PRIMARY KEY,
    post_id text,
    mime text NOT NULL,
    s3_key text NOT NULL,
    created_at text NOT NULL,
    deleted_at text,
    FOREIGN KEY (post_id) REFERENCES posts(id)
);

CREATE UNIQUE INDEX idx_uniq_media_s3_key ON media (s3_key);
