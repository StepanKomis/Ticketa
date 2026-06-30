-- name: GetServerSetting :one
SELECT key, value, from_env, updated_at
FROM server_settings
WHERE key = $1;

-- name: GetAllServerSettings :many
SELECT key, value, from_env, updated_at
FROM server_settings
ORDER BY key;

-- name: UpsertServerSetting :exec
INSERT INTO server_settings (key, value, from_env, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (key) DO UPDATE
    SET value      = EXCLUDED.value,
        from_env   = EXCLUDED.from_env,
        updated_at = NOW();
