CREATE TABLE tickets (
    id         BIGSERIAL   PRIMARY KEY,
    title      TEXT        NOT NULL,
    body       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    author_id  INTEGER     NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status_id  INTEGER     REFERENCES ticket_statuses(id) ON DELETE SET NULL
);

CREATE INDEX idx_tickets_author_id ON tickets (author_id);
CREATE INDEX idx_tickets_status_id ON tickets (status_id);
