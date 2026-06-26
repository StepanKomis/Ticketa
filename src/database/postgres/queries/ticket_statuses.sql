-- name: CreateTicketStatus :one
INSERT INTO ticket_statuses (title, color, position, is_closed)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTicketStatus :one
SELECT * FROM ticket_statuses
WHERE id = $1;

-- name: ListTicketStatuses :many
SELECT * FROM ticket_statuses
ORDER BY position ASC;

-- name: UpsertTicketStatusByPosition :one
INSERT INTO ticket_statuses (title, color, position, is_closed)
VALUES ($1, $2, $3, $4)
ON CONFLICT (position) DO UPDATE
    SET title = EXCLUDED.title,
        color = EXCLUDED.color,
        is_closed = EXCLUDED.is_closed
RETURNING *;

-- name: UpdateTicketStatus :one
UPDATE ticket_statuses
SET title = $2,
    color = $3,
    is_closed = $4
WHERE id = $1
RETURNING *;

-- name: DeleteTicketStatus :exec
DELETE FROM ticket_statuses
WHERE id = $1;
