-- name: GetFeeds :many

SELECT 
    f.name  AS feed_name,
    f.url   AS feed_url,
    u.name  AS user_name
FROM feeds AS f
JOIN users AS u ON f.user_id = u.id;