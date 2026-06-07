CREATE TABLE ticket_statuses (
    id       SERIAL       PRIMARY KEY,
    title    VARCHAR(100) NOT NULL,
    color    CHAR(7)      NOT NULL DEFAULT '#808080',
    position INTEGER      NOT NULL,
    CONSTRAINT uq_ticket_statuses_position UNIQUE (position)
);
