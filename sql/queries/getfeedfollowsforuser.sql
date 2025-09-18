-- name: GetFeedFollowsForUser :many

SELECT 
    feeds.name AS feed_name,
    users.name AS user_name
FROM feed_follow
INNER JOIN feeds ON feed_follow.feed_id = feeds.id
INNER JOIN users ON feed_follow.user_id = users.id
WHERE users.id = $1;