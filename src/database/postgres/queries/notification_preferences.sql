-- name: GetEmailOptOuts :many
SELECT type FROM notification_email_optouts WHERE user_id = $1;

-- name: UpsertEmailOptOut :exec
INSERT INTO notification_email_optouts (user_id, type)
VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: DeleteEmailOptOut :exec
DELETE FROM notification_email_optouts WHERE user_id = $1 AND type = $2;

-- name: GetAllActiveStaff :many
SELECT id, email, first_name, last_name FROM users
WHERE user_type IN ('staff', 'maintainer', 'admin') AND is_active = TRUE;
