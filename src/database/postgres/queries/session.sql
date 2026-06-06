-- name: GetSessionByToken :one
UPDATE sessions
SET    last_seen_at = NOW()
WHERE  token      = $1
  AND  deleted    = FALSE
  AND  expires_at > NOW()
RETURNING *;

-- name: GetSessionByUserID :one
SELECT * FROM sessions
WHERE user_id = $1
LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, token, ip, user_agent, expires_at)
VALUES ($1, $2, $3, $4, NOW() + $5::interval)
RETURNING *;

-- name: RegenerateToken :one
UPDATE sessions
SET    token        = $2,
       ip           = $3,
       user_agent   = $4,
       expires_at   = NOW() + $5::interval,
       last_seen_at = NOW(),
       deleted      = FALSE
WHERE  user_id = $1
RETURNING *;

-- name: SoftDeleteSession :exec
UPDATE sessions
SET    deleted    = TRUE,
       token      = '',
       expires_at = NOW()
WHERE  token   = $1
  AND  deleted = FALSE;

-- name: SoftDeleteSessionByUserID :exec
UPDATE sessions
SET    deleted    = TRUE,
       token      = '',
       expires_at = NOW()
WHERE  user_id = $1
  AND  deleted = FALSE;