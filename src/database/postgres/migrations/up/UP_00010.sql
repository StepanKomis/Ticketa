CREATE TABLE ticket_history (
    id         BIGSERIAL    PRIMARY KEY,
    ticket_id  BIGINT       NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    actor_id   INTEGER      NOT NULL REFERENCES users(id)  ON DELETE RESTRICT,
    actor_name TEXT         NOT NULL DEFAULT '',
    event      VARCHAR(30)  NOT NULL,
    old_val    TEXT,
    new_val    TEXT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ticket_history_ticket ON ticket_history(ticket_id, created_at);
