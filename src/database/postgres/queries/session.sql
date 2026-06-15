-- name: GetSessionByToken :one
-- Session je platná jen pokud je její uživatel stále aktivní — deaktivace účtu
-- tak okamžitě zneplatní všechny jeho požadavky.
UPDATE sessions
SET    last_seen_at = NOW()
FROM   users
WHERE  sessions.token      = $1
  AND  sessions.deleted    = FALSE
  AND  sessions.expires_at > NOW()
  AND  users.id            = sessions.user_id
  AND  users.is_active     = TRUE
RETURNING sessions.*;

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