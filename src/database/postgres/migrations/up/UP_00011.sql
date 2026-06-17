CREATE TABLE activity_log (
    id          BIGSERIAL    PRIMARY KEY,
    event_type  VARCHAR(50)  NOT NULL,
    actor_id    INTEGER      REFERENCES users(id) ON DELETE SET NULL,
    target_type VARCHAR(50),
    target_id   BIGINT,
    payload     JSONB,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_activity_log_actor  ON activity_log(actor_id, created_at);
CREATE INDEX idx_activity_log_target ON activity_log(target_type, target_id, created_at);
