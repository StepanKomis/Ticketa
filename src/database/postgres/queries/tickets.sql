-- name: CreateTicket :one
INSERT INTO tickets (title, body, author_id, status_id, priority, location, category, assigned_to, requested_priority)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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
    AND (sqlc.narg('pending_priority_approval')::BOOLEAN IS NULL OR (t.requested_priority IS NOT NULL) = sqlc.narg('pending_priority_approval'))
    AND (sqlc.narg('unassigned_only')::BOOLEAN IS NULL OR (t.assigned_to IS NULL) = sqlc.narg('unassigned_only'))
    AND (sqlc.narg('closed')::BOOLEAN IS NULL OR t.is_closed = sqlc.narg('closed'))
    AND (
        CASE WHEN sqlc.narg('show_deleted')::BOOLEAN IS NOT DISTINCT FROM TRUE
            THEN t.deleted_at IS NOT NULL
            ELSE t.deleted_at IS NULL
        END
    )
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
    AND (sqlc.narg('pending_priority_approval')::BOOLEAN IS NULL OR (t.requested_priority IS NOT NULL) = sqlc.narg('pending_priority_approval'))
    AND (sqlc.narg('unassigned_only')::BOOLEAN IS NULL OR (t.assigned_to IS NULL) = sqlc.narg('unassigned_only'))
    AND (sqlc.narg('closed')::BOOLEAN IS NULL OR t.is_closed = sqlc.narg('closed'))
    AND (
        CASE WHEN sqlc.narg('show_deleted')::BOOLEAN IS NOT DISTINCT FROM TRUE
            THEN t.deleted_at IS NOT NULL
            ELSE t.deleted_at IS NULL
        END
    )
    AND (
        sqlc.arg('q')::TEXT = ''
        OR t.title ILIKE '%' || sqlc.arg('q') || '%'
        OR t.body  ILIKE '%' || sqlc.arg('q') || '%'
    );

-- name: UpdateTicket :one
-- touch_status_id rozlišuje "status_id v requestu nebyl uveden" (status_id se
-- nemění) od "status_id byl explicitně poslán" (i jako null) — bez toho by
-- každá úprava title/body bez status_id vynulovala stav tiketu.
UPDATE tickets
SET title               = COALESCE(sqlc.narg('title'),    title),
    body                = COALESCE(sqlc.narg('body'),     body),
    priority            = COALESCE(sqlc.narg('priority'), priority),
    location            = COALESCE(sqlc.narg('location'), location),
    category            = COALESCE(sqlc.narg('category'), category),
    status_id           = CASE WHEN sqlc.arg('touch_status_id')::boolean THEN sqlc.narg('status_id') ELSE status_id END,
    requested_priority  = COALESCE(sqlc.narg('requested_priority'), requested_priority),
    resolution_note     = COALESCE(sqlc.narg('resolution_note'), resolution_note)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdateTicketMeta :one
-- touch_assigned_to/touch_status_id: stejný důvod jako u UpdateTicket výše —
-- PATCH s jen jedním polem nesmí vynulovat to druhé.
UPDATE tickets
SET assigned_to     = CASE WHEN sqlc.arg('touch_assigned_to')::boolean THEN sqlc.narg('assigned_to') ELSE assigned_to END,
    status_id       = CASE WHEN sqlc.arg('touch_status_id')::boolean   THEN sqlc.narg('status_id')   ELSE status_id   END,
    priority        = COALESCE(sqlc.narg('priority'), priority),
    location        = COALESCE(sqlc.narg('location'), location),
    category        = COALESCE(sqlc.narg('category'), category),
    resolution_note = COALESCE(sqlc.narg('resolution_note'), resolution_note)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: ApproveTicketPriority :one
UPDATE tickets
SET priority             = requested_priority,
    priority_approved_by = sqlc.arg('approved_by')::INTEGER,
    requested_priority   = NULL
WHERE id = sqlc.arg('id') AND requested_priority IS NOT NULL
RETURNING *;

-- name: RejectTicketPriority :one
UPDATE tickets
SET requested_priority = NULL
WHERE id = sqlc.arg('id') AND requested_priority IS NOT NULL
RETURNING *;

-- name: DeleteTicket :exec
DELETE FROM tickets
WHERE id = $1;

-- name: SoftDeleteTicket :exec
UPDATE tickets
SET deleted_at = NOW()
WHERE id = $1;

-- name: VoteTicket :exec
INSERT INTO ticket_votes (ticket_id, user_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnvoteTicket :exec
DELETE FROM ticket_votes WHERE ticket_id = $1 AND user_id = $2;

-- name: InsertTicketHistory :exec
INSERT INTO ticket_history (ticket_id, actor_id, actor_name, event, old_val, new_val)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListTicketHistory :many
SELECT id, ticket_id, actor_id, actor_name, event, old_val, new_val, created_at
FROM ticket_history
WHERE ticket_id = $1
ORDER BY created_at ASC
LIMIT  sqlc.arg('lim')::INTEGER
OFFSET sqlc.arg('off')::INTEGER;

-- name: GetStatusTitle :one
SELECT title FROM ticket_statuses WHERE id = $1;
