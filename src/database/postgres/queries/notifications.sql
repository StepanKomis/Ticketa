-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, text, ticket_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetNotificationsForUser :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 50;

-- name: CountUnreadNotifications :one
SELECT COUNT(*)::BIGINT
FROM notifications
WHERE user_id = $1 AND is_viewed = FALSE;

-- name: MarkAllNotificationsViewed :exec
UPDATE notifications
SET is_viewed = TRUE
WHERE user_id = $1 AND is_viewed = FALSE;
