ALTER TABLE tickets
    ADD COLUMN priority    VARCHAR(10)  NOT NULL DEFAULT 'medium'
                           CHECK (priority IN ('low','medium','high','urgent')),
    ADD COLUMN assigned_to INTEGER      REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN location    VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN category    VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW();

CREATE FUNCTION set_ticket_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$;

CREATE TRIGGER trg_ticket_updated_at
BEFORE UPDATE ON tickets
FOR EACH ROW EXECUTE FUNCTION set_ticket_updated_at();

CREATE INDEX idx_tickets_assigned_to ON tickets(assigned_to);
CREATE INDEX idx_tickets_status_priority ON tickets(status_id, priority);

CREATE TABLE ticket_votes (
    ticket_id  BIGINT  NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (ticket_id, user_id)
);

CREATE INDEX idx_ticket_votes_ticket ON ticket_votes(ticket_id);
