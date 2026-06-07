-- name: CreateTicket :one
INSERT INTO tickets (title, body, author_id, status_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTicket :one
SELECT * FROM tickets
WHERE id = $1;

-- name: ListTickets :many
SELECT * FROM tickets
ORDER BY created_at DESC;

-- name: UpdateTicket :one
UPDATE tickets
SET title     = $2,
    body      = $3,
    status_id = $4
WHERE id = $1
RETURNING *;

-- name: DeleteTicket :exec
DELETE FROM tickets
WHERE id = $1;
