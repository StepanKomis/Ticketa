-- name: GetUserByID :one
SELECT id, email, first_name, last_name, user_type, provider, is_active, created_at, last_login_at, requested_role, approved_by
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, first_name, last_name, user_type, provider, is_active, created_at, last_login_at, requested_role, approved_by
FROM users
WHERE email = $1;

-- name: GetUserIDByEmail :one
SELECT id FROM users WHERE email = $1;

-- name: ListUsers :many
SELECT id, email, first_name, last_name, user_type, provider, is_active, created_at, last_login_at, requested_role, approved_by
FROM users
ORDER BY created_at DESC;

-- name: ListUsersByType :many
SELECT id, email, first_name, last_name, user_type, provider, is_active, created_at, last_login_at
FROM users
WHERE user_type = $1
ORDER BY created_at DESC;

-- name: ListActiveUsers :many
SELECT id, email, first_name, last_name, user_type, provider, is_active, created_at, last_login_at
FROM users
WHERE is_active = TRUE
ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (email, first_name, last_name, user_type, provider)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, first_name, last_name, user_type, provider, is_active, created_at, last_login_at, requested_role, approved_by;

-- name: UpdateUser :one
UPDATE users
SET email      = $2,
    first_name = $3,
    last_name  = $4
WHERE id = $1
RETURNING id, email, first_name, last_name, user_type, provider, is_active, created_at, last_login_at, requested_role, approved_by;

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users
SET is_active = FALSE
WHERE id = $1;

-- name: SetUserIsActive :exec
UPDATE users
SET is_active = $2
WHERE id = $1;

-- name: SetUserType :one
UPDATE users
SET user_type = $2
WHERE id = $1
RETURNING id, email, first_name, last_name, user_type, provider, is_active, created_at, last_login_at, requested_role, approved_by;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetStudentByID :one
SELECT
    u.id, u.email, u.first_name, u.last_name, u.provider, u.is_active, u.created_at, u.last_login_at,
    sp.student_id, sp.enrolled_at, sp.graduation_at, sp.class_group
FROM users u
JOIN student_profile sp ON sp.id = u.id
WHERE u.id = $1;

-- name: GetStaffByID :one
SELECT
    u.id, u.email, u.first_name, u.last_name, u.provider, u.is_active, u.created_at, u.last_login_at,
    sfp.employee_id, sfp.department, sfp.job_title, sfp.is_teacher, sfp.employed_since
FROM users u
JOIN staff_profile sfp ON sfp.id = u.id
WHERE u.id = $1;

-- name: GetMaintainerByID :one
SELECT
    u.id, u.email, u.first_name, u.last_name, u.provider, u.is_active, u.created_at, u.last_login_at,
    mp.access_level, mp.managed_scope
FROM users u
JOIN maintainer_profile mp ON mp.id = u.id
WHERE u.id = $1;

-- name: GetUserWithLocalLogin :one
SELECT
    u.id, u.email, u.first_name, u.last_name, u.user_type, u.is_active, u.created_at, u.last_login_at,
    ll.password_hash, ll.must_change_pw, ll.pw_changed_at
FROM users u
JOIN local_login ll ON ll.id = u.id
WHERE u.email = $1 AND u.is_active = TRUE;

-- name: GetUserWithLDAPLogin :one
SELECT
    u.id, u.email, u.first_name, u.last_name, u.user_type, u.provider, u.is_active, u.created_at, u.last_login_at,
    ldap.distinguished_name, ldap.uid, ldap.sam_account_name, ldap.upn,
    ldap.object_guid, ldap.object_sid, ldap.ldap_server, ldap.base_dn, ldap.last_synced_at
FROM users u
JOIN ldap_login ldap ON ldap.id = u.id
WHERE ldap.distinguished_name = $1 AND u.is_active = TRUE;

-- name: CreateLocalLogin :exec
INSERT INTO local_login (id, password_hash, must_change_pw)
VALUES ($1, $2, $3);

-- name: UpdateLocalLoginPassword :exec
UPDATE local_login
SET password_hash  = $2,
    must_change_pw = FALSE,
    pw_changed_at  = NOW()
WHERE id = $1;

-- name: CreateLDAPLogin :one
INSERT INTO ldap_login (id, distinguished_name, uid, sam_account_name, upn, object_guid, object_sid, ldap_server, base_dn)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, distinguished_name, uid, sam_account_name, upn, object_guid, object_sid, ldap_server, base_dn, last_synced_at;

-- name: SyncLDAPLogin :exec
UPDATE ldap_login
SET uid              = $2,
    sam_account_name = $3,
    upn              = $4,
    object_guid      = $5,
    object_sid       = $6,
    last_synced_at   = NOW()
WHERE id = $1;

-- name: CreateStudentProfile :one
INSERT INTO student_profile (id, student_id, enrolled_at, graduation_at, class_group)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, student_id, enrolled_at, graduation_at, class_group;

-- name: UpdateStudentProfile :exec
UPDATE student_profile
SET graduation_at = $2,
    class_group   = $3
WHERE id = $1;

-- name: CreateStaffProfile :one
INSERT INTO staff_profile (id, employee_id, department, job_title, is_teacher, employed_since)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, employee_id, department, job_title, is_teacher, employed_since;

-- name: UpdateStaffProfile :exec
UPDATE staff_profile
SET department = $2,
    job_title  = $3,
    is_teacher = $4
WHERE id = $1;

-- name: CreateMaintainerProfile :one
INSERT INTO maintainer_profile (id, access_level, managed_scope)
VALUES ($1, $2, $3)
RETURNING id, access_level, managed_scope;

-- name: UpdateMaintainerProfile :exec
UPDATE maintainer_profile
SET access_level  = $2,
    managed_scope = $3
WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: SetRequestedRole :exec
UPDATE users SET requested_role = $2 WHERE id = $1;

-- name: GetPendingUsers :many
SELECT id, email, first_name, last_name, user_type, provider, is_active, created_at, last_login_at, requested_role, approved_by
FROM users WHERE user_type = 'pending' ORDER BY created_at DESC;

-- name: ApprovePendingUser :exec
UPDATE users SET user_type = requested_role, approved_by = $2 WHERE id = $1;

-- name: RejectPendingUser :exec
UPDATE users SET is_active = FALSE WHERE id = $1;

-- name: UpdateMyProfile :one
UPDATE users
SET first_name = COALESCE($2, first_name),
    last_name  = COALESCE($3, last_name)
WHERE id = $1
RETURNING id, email, first_name, last_name, user_type, provider, is_active, created_at, last_login_at;
