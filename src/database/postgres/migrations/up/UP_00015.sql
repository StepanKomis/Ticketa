ALTER TABLE tickets ADD COLUMN resolved_at TIMESTAMPTZ;

-- Backfill existujících uzavřených tiketů
UPDATE tickets SET resolved_at = updated_at WHERE is_closed = TRUE;

-- Trigger: nastaví/vymaže resolved_at při přechodu is_closed.
-- Spouští se po trg_ticket_is_closed (alphabetically 'r' > 'i'),
-- takže NEW.is_closed je již nastaveno správnou hodnotou z prvního triggeru.
CREATE FUNCTION set_ticket_resolved_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.is_closed = TRUE AND NOT COALESCE(OLD.is_closed, FALSE) THEN
        NEW.resolved_at := NOW();
    ELSIF NOT NEW.is_closed AND COALESCE(OLD.is_closed, FALSE) THEN
        NEW.resolved_at := NULL;
    END IF;
    RETURN NEW;
END;
$$;

-- UPDATE OF status_id: normální změna statusu (trg_ticket_is_closed běží dřív)
-- UPDATE OF is_closed: admin překonfiguroval který status je "uzavřený"
CREATE TRIGGER trg_ticket_resolved_at
BEFORE INSERT OR UPDATE OF status_id, is_closed ON tickets
FOR EACH ROW EXECUTE FUNCTION set_ticket_resolved_at();
