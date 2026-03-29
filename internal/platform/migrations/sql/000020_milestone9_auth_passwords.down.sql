ALTER TABLE identityaccess.users
DROP CONSTRAINT IF EXISTS identityaccess_users_password_hash_not_blank;

ALTER TABLE identityaccess.users
DROP COLUMN IF EXISTS password_updated_at,
DROP COLUMN IF EXISTS password_hash;
