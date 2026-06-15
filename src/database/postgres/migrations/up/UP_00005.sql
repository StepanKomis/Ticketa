CREATE TABLE ticket_comments (
    id         BIGSERIAL   PRIMARY KEY,
    ticket_id  BIGINT      NOT NULL REFERENCES tickets(id)         ON DELETE CASCADE,
    author_id  INTEGER     NOT NULL REFERENCES users(id)           ON DELETE RESTRICT,
    parent_id  BIGINT               REFERENCES ticket_comments(id) ON DELETE SET NULL,
    body       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted    BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_ticket_comments_ticket_id ON ticket_comments (ticket_id) WHERE deleted = FALSE;
CREATE INDEX idx_ticket_comments_parent_id ON ticket_comments (parent_id);
