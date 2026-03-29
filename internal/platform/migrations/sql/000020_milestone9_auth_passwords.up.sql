ALTER TABLE identityaccess.users
ADD COLUMN password_hash TEXT,
ADD COLUMN password_updated_at TIMESTAMPTZ;

ALTER TABLE identityaccess.users
ADD CONSTRAINT identityaccess_users_password_hash_not_blank
CHECK (password_hash IS NULL OR btrim(password_hash) <> '');
