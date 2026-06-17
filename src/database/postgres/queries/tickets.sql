-- name: CreateTicket :one
INSERT INTO tickets (title, body, author_id, status_id, priority, location, category, assigned_to)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetTicket :one
SELECT
    t.*,
    COALESCE(u.first_name || ' ' || u.last_name, u.email)     AS author_name,
    COALESCE(a.first_name || ' ' || a.last_name, a.email, '') AS assignee_name,
    (SELECT COUNT(*)::INT FROM ticket_votes tv WHERE tv.ticket_id = t.id) AS vote_count,
    EXISTS(
        SELECT 1 FROM ticket_votes tv
        WHERE tv.ticket_id = t.id AND tv.user_id = sqlc.arg('current_user_id')::INTEGER
    ) AS user_has_voted
FROM tickets t
JOIN  users u ON u.id = t.author_id
LEFT JOIN users a ON a.id = t.assigned_to
WHERE t.id = sqlc.arg('id')::BIGINT;

-- name: ListTicketsFiltered :many
SELECT
    t.*,
    COALESCE(u.first_name || ' ' || u.last_name, u.email)     AS author_name,
    COALESCE(a.first_name || ' ' || a.last_name, a.email, '') AS assignee_name,
    (SELECT COUNT(*)::INT FROM ticket_votes tv WHERE tv.ticket_id = t.id) AS vote_count,
    EXISTS(
        SELECT 1 FROM ticket_votes tv
        WHERE tv.ticket_id = t.id AND tv.user_id = sqlc.arg('current_user_id')::INTEGER
    ) AS user_has_voted
FROM tickets t
JOIN  users u ON u.id = t.author_id
LEFT JOIN users a ON a.id = t.assigned_to
WHERE
    (sqlc.narg('status_id')::INTEGER IS NULL    OR t.status_id   = sqlc.narg('status_id'))
    AND (sqlc.narg('priority')::VARCHAR IS NULL  OR t.priority    = sqlc.narg('priority'))
    AND (sqlc.narg('assigned_to')::INTEGER IS NULL OR t.assigned_to = sqlc.narg('assigned_to'))
    AND (sqlc.narg('author_id')::INTEGER IS NULL OR t.author_id   = sqlc.narg('author_id'))
    AND (sqlc.narg('category')::VARCHAR IS NULL  OR t.category    = sqlc.narg('category'))
    AND (
        sqlc.arg('q')::TEXT = ''
        OR t.title ILIKE '%' || sqlc.arg('q') || '%'
        OR t.body  ILIKE '%' || sqlc.arg('q') || '%'
    )
ORDER BY t.created_at DESC
LIMIT  sqlc.arg('lim')::INTEGER
OFFSET sqlc.arg('off')::INTEGER;

-- name: CountTicketsFiltered :one
SELECT COUNT(*)::BIGINT
FROM tickets t
WHERE
    (sqlc.narg('status_id')::INTEGER IS NULL    OR t.status_id   = sqlc.narg('status_id'))
    AND (sqlc.narg('priority')::VARCHAR IS NULL  OR t.priority    = sqlc.narg('priority'))
    AND (sqlc.narg('assigned_to')::INTEGER IS NULL OR t.assigned_to = sqlc.narg('assigned_to'))
    AND (sqlc.narg('author_id')::INTEGER IS NULL OR t.author_id   = sqlc.narg('author_id'))
    AND (sqlc.narg('category')::VARCHAR IS NULL  OR t.category    = sqlc.narg('category'))
    AND (
        sqlc.arg('q')::TEXT = ''
        OR t.title ILIKE '%' || sqlc.arg('q') || '%'
        OR t.body  ILIKE '%' || sqlc.arg('q') || '%'
    );

-- name: UpdateTicket :one
UPDATE tickets
SET title     = COALESCE(sqlc.narg('title'),    title),
    body      = COALESCE(sqlc.narg('body'),     body),
    priority  = COALESCE(sqlc.narg('priority'), priority),
    location  = COALESCE(sqlc.narg('location'), location),
    category  = COALESCE(sqlc.narg('category'), category),
    status_id = sqlc.narg('status_id')
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdateTicketMeta :one
UPDATE tickets
SET assigned_to = sqlc.narg('assigned_to'),
    status_id   = sqlc.narg('status_id'),
    priority    = COALESCE(sqlc.narg('priority'), priority),
    location    = COALESCE(sqlc.narg('location'), location),
    category    = COALESCE(sqlc.narg('category'), category)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteTicket :exec
DELETE FROM tickets
WHERE id = $1;

-- name: VoteTicket :exec
INSERT INTO ticket_votes (ticket_id, user_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnvoteTicket :exec
DELETE FROM ticket_votes WHERE ticket_id = $1 AND user_id = $2;
