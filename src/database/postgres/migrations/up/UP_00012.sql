ALTER TABLE tickets
    ADD COLUMN requested_priority VARCHAR(10)
        CHECK (requested_priority IS NULL OR requested_priority IN ('low', 'medium', 'high', 'urgent')),
    ADD COLUMN priority_approved_by INTEGER REFERENCES users(id);
