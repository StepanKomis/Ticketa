ALTER TABLE users
  ADD COLUMN requested_role user_type NULL,
  ADD COLUMN approved_by    INTEGER   REFERENCES users(id) ON DELETE SET NULL NULL;
