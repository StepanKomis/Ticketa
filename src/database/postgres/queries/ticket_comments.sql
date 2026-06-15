-- name: CreateComment :one
INSERT INTO ticket_comments (ticket_id, author_id, parent_id, body)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetComment :one
SELECT * FROM ticket_comments WHERE id = $1;

-- name: ListCommentsByTicket :many
SELECT * FROM ticket_comments
WHERE ticket_id = $1 AND deleted = FALSE
ORDER BY created_at ASC;

-- name: UpdateComment :one
UPDATE ticket_comments
SET body = $2, updated_at = NOW()
WHERE id = $1 AND deleted = FALSE
RETURNING *;

-- name: SoftDeleteComment :exec
UPDATE ticket_comments
SET deleted = TRUE, updated_at = NOW()
WHERE id = $1;
