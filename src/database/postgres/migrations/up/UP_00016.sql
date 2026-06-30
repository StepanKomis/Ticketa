-- Soft delete pro tikety
ALTER TABLE tickets ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_tickets_deleted_at ON tickets(deleted_at)
    WHERE deleted_at IS NOT NULL;

-- Tabulka oznámení
CREATE TABLE notifications (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    INTEGER      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       VARCHAR(64)  NOT NULL,
    text       TEXT         NOT NULL,
    ticket_id  BIGINT       REFERENCES tickets(id),
    is_viewed  BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_unread
    ON notifications(user_id, is_viewed)
    WHERE is_viewed = FALSE;
