CREATE TABLE server_settings (
    key        TEXT        PRIMARY KEY,
    value      TEXT        NOT NULL,
    from_env   BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO server_settings (key, value, from_env)
VALUES ('wizard_completed', 'false', false);
