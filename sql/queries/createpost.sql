-- name: CreatePost :exec
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
ON CONFLICT (url)
DO UPDATE SET
    updated_at = EXCLUDED.updated_at,
    title = EXCLUDED.title,
    published_at = EXCLUDED.published_at,
    feed_id = EXCLUDED.feed_id
WHERE posts.published_at < EXCLUDED.published_at;