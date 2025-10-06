-- name: GetPostsforUser :many

SELECT posts.*
FROM users 
LEFT JOIN feed_follow ON users.id = feed_follow.user_id
LEFT JOIN feeds ON feed_follow.feed_id = feeds.id
LEFT JOIN posts ON feeds.id = posts.feed_id
WHERE users.name = $1
ORDER BY published_at DESC
LIMIT $2;