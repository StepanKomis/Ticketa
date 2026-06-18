-- name: CreateActivityLog :exec
INSERT INTO activity_log (event_type, actor_id, target_type, target_id, payload)
VALUES ($1, $2, $3, $4, $5);

-- name: ListActivityLog :many
SELECT * FROM activity_log
WHERE
    (sqlc.narg('event_type')::VARCHAR IS NULL  OR event_type  = sqlc.narg('event_type'))
    AND (sqlc.narg('actor_id')::INTEGER IS NULL OR actor_id    = sqlc.narg('actor_id'))
    AND (sqlc.narg('target_type')::VARCHAR IS NULL OR target_type = sqlc.narg('target_type'))
    AND (sqlc.narg('target_id')::BIGINT IS NULL OR target_id   = sqlc.narg('target_id'))
    AND (sqlc.narg('from_ts')::TIMESTAMPTZ IS NULL OR created_at >= sqlc.narg('from_ts'))
    AND (sqlc.narg('to_ts')::TIMESTAMPTZ IS NULL   OR created_at <= sqlc.narg('to_ts'))
ORDER BY created_at DESC
LIMIT  sqlc.arg('lim')::INTEGER
OFFSET sqlc.arg('off')::INTEGER;

-- name: CountActivityLog :one
SELECT COUNT(*) FROM activity_log
WHERE
    (sqlc.narg('event_type')::VARCHAR IS NULL  OR event_type  = sqlc.narg('event_type'))
    AND (sqlc.narg('actor_id')::INTEGER IS NULL OR actor_id    = sqlc.narg('actor_id'))
    AND (sqlc.narg('target_type')::VARCHAR IS NULL OR target_type = sqlc.narg('target_type'))
    AND (sqlc.narg('target_id')::BIGINT IS NULL OR target_id   = sqlc.narg('target_id'))
    AND (sqlc.narg('from_ts')::TIMESTAMPTZ IS NULL OR created_at >= sqlc.narg('from_ts'))
    AND (sqlc.narg('to_ts')::TIMESTAMPTZ IS NULL   OR created_at <= sqlc.narg('to_ts'));

-- name: ListActivityLogForUser :many
-- Zahrnuje vlastní akce uživatele a dále akce ostatních na tiketech, jejichž
-- je uživatel autorem (např. když štáb změní stav mého tiketu).
SELECT * FROM activity_log
WHERE actor_id = sqlc.arg('actor_id')::INTEGER
   OR (target_type = 'ticket' AND target_id IN (
         SELECT id FROM tickets WHERE author_id = sqlc.arg('actor_id')::INTEGER
       ))
ORDER BY created_at DESC
LIMIT  sqlc.arg('lim')::INTEGER
OFFSET sqlc.arg('off')::INTEGER;

-- name: CountActivityLogForUser :one
SELECT COUNT(*) FROM activity_log
WHERE actor_id = sqlc.arg('actor_id')::INTEGER
   OR (target_type = 'ticket' AND target_id IN (
         SELECT id FROM tickets WHERE author_id = sqlc.arg('actor_id')::INTEGER
       ));
