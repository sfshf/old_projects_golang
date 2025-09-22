USE slark;

ALTER TABLE slk_user ADD COLUMN secondary_password_hash VARCHAR(64) NULL COMMENT 'secondary password hash' AFTER password_hash;