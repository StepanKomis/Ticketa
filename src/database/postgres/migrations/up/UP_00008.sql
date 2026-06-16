CREATE TABLE invitations (
  id         BIGSERIAL    PRIMARY KEY,
  email      VARCHAR(255) UNIQUE NOT NULL,
  invited_by INTEGER      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  token      CHAR(64)     UNIQUE NOT NULL,
  user_type  user_type    NOT NULL,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ  NOT NULL,
  used_at    TIMESTAMPTZ  NULL
);
