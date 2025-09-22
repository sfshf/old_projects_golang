USE slark;

ALTER TABLE slk_user MODIFY password_hash VARCHAR(64) NULL COMMENT 'password hash';