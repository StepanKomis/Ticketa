-- name: CreateInvitation :one
INSERT INTO invitations (email, invited_by, token, user_type, expires_at)
VALUES ($1, $2, $3, $4, NOW() + INTERVAL '7 days')
RETURNING id, email, invited_by, token, user_type, created_at, expires_at, used_at;

-- name: GetInvitationByToken :one
SELECT id, email, invited_by, token, user_type, created_at, expires_at, used_at
FROM invitations
WHERE token = $1;

-- name: MarkInvitationUsed :exec
UPDATE invitations SET used_at = NOW() WHERE id = $1;
