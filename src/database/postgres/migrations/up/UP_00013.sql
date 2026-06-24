ALTER TABLE ticket_statuses ADD COLUMN is_closed BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE tickets         ADD COLUMN is_closed BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE tickets         ADD COLUMN resolution_note TEXT;
CREATE INDEX idx_tickets_is_closed ON tickets(is_closed);

-- Denormalizovaný tickets.is_closed se odvozuje ze ticket_statuses.is_closed
-- při každém vložení/změně status_id — DB trigger, ne aplikační kód, aby to
-- nešlo zapomenout v některém z handlerů (assign/patch/update/claim/create).
CREATE FUNCTION sync_ticket_is_closed()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.status_id IS NULL THEN
        NEW.is_closed := FALSE;
    ELSE
        SELECT is_closed INTO NEW.is_closed FROM ticket_statuses WHERE id = NEW.status_id;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_ticket_is_closed
BEFORE INSERT OR UPDATE OF status_id ON tickets
FOR EACH ROW EXECUTE FUNCTION sync_ticket_is_closed();

-- Když admin překonfiguruje, který status je "uzavřený", existující tikety
-- v tom stavu se musí dotáhnout zpětně, jinak by jim is_closed zůstalo staré.
CREATE FUNCTION sync_tickets_is_closed_on_status_change()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    UPDATE tickets SET is_closed = NEW.is_closed WHERE status_id = NEW.id;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_ticket_statuses_is_closed
AFTER UPDATE OF is_closed ON ticket_statuses
FOR EACH ROW EXECUTE FUNCTION sync_tickets_is_closed_on_status_change();
