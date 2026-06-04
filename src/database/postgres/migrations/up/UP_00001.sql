-- +migrate Up

CREATE TYPE auth_provider AS ENUM ('local', 'ldap', 'ad');
CREATE TYPE user_type     AS ENUM ('student', 'staff', 'maintainer');

CREATE TABLE users (
    id            SERIAL        PRIMARY KEY,
    email         VARCHAR(255)  NOT NULL,
    first_name    VARCHAR(60),
    last_name     VARCHAR(60),
    user_type     user_type     NOT NULL,
    provider      auth_provider NOT NULL DEFAULT 'local',
    is_active     BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_email ON users (email);

-- auth: local password login (provider = 'local')
CREATE TABLE local_login (
    id             INTEGER      PRIMARY KEY REFERENCES users (id),
    password_hash  VARCHAR(255) NOT NULL,
    must_change_pw BOOLEAN      NOT NULL DEFAULT FALSE,
    pw_changed_at  TIMESTAMPTZ
);

-- auth: LDAP / Active Directory login (provider = 'ldap' or 'ad')
CREATE TABLE ldap_login (
    id                 INTEGER      PRIMARY KEY REFERENCES users (id),
    distinguished_name VARCHAR(512) NOT NULL UNIQUE,
    uid                VARCHAR(255),
    sam_account_name   VARCHAR(255),
    upn                VARCHAR(255),
    object_guid        UUID         UNIQUE,
    object_sid         VARCHAR(184),
    ldap_server        VARCHAR(255) NOT NULL,
    base_dn            VARCHAR(512) NOT NULL,
    last_synced_at     TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_ldap_dn   ON ldap_login (distinguished_name);
CREATE UNIQUE INDEX idx_ldap_guid ON ldap_login (object_guid);
CREATE        INDEX idx_ldap_sam  ON ldap_login (sam_account_name);
CREATE        INDEX idx_ldap_upn  ON ldap_login (upn);

-- profile: students
CREATE TABLE student_profile (
    id            INTEGER     PRIMARY KEY REFERENCES users (id),
    student_id    VARCHAR(50) NOT NULL UNIQUE,
    enrolled_at   DATE        NOT NULL,
    graduation_at DATE,
    class_group   VARCHAR(60)
);

CREATE UNIQUE INDEX idx_student_id ON student_profile (student_id);

-- profile: staff (teachers + other workers)
CREATE TABLE staff_profile (
    id             INTEGER      PRIMARY KEY REFERENCES users (id),
    employee_id    VARCHAR(50)  NOT NULL UNIQUE,
    department     VARCHAR(100),
    job_title      VARCHAR(100),
    is_teacher     BOOLEAN      NOT NULL DEFAULT FALSE,
    employed_since DATE
);

CREATE UNIQUE INDEX idx_employee_id ON staff_profile (employee_id);

-- profile: maintainers
CREATE TABLE maintainer_profile (
    id            INTEGER      PRIMARY KEY REFERENCES users (id),
    access_level  SMALLINT     NOT NULL DEFAULT 1,
    managed_scope VARCHAR(255)
);
