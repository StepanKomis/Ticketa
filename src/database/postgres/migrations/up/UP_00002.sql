-- Migration: 000002_create_sessions
-- Depends on: users table with users.id (BIGINT or UUID — adjust FK type below)
 
CREATE TABLE IF NOT EXISTS sessions (
    id           BIGSERIAL                  PRIMARY KEY,
    user_id      BIGINT                     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 
    token        CHAR(64)                   NOT NULL DEFAULT '',
    ip           INET                       NOT NULL,
    user_agent   TEXT                       NOT NULL,
 
    created_at   TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ                NOT NULL,
    last_seen_at TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
 
    -- soft delete: true = session is invalidated, token is cleared
    deleted      BOOLEAN                    NOT NULL DEFAULT FALSE,
 
    -- one active session per user
    CONSTRAINT uq_sessions_user_id UNIQUE (user_id)
);
 
CREATE INDEX idx_sessions_token     ON sessions (token) WHERE deleted = FALSE;
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at) WHERE deleted = FALSE;
 

 
