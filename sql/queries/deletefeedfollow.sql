-- name: DeleteFeedFollow :exec

DELETE FROM feed_follow WHERE user_id = $1 AND feed_id = $2;