USE slark;

ALTER TABLE slk_user DROP COLUMN first_name;
ALTER TABLE slk_user DROP COLUMN last_name;
ALTER TABLE slk_user MODIFY email VARCHAR(64) NULL COMMENT 'email address';
ALTER TABLE slk_user MODIFY phone VARCHAR(16) NULL COMMENT 'phone number';
ALTER TABLE slk_user DROP COLUMN gender;
ALTER TABLE slk_user DROP COLUMN birthday;

ALTER TABLE slk_user MODIFY COLUMN nickname VARCHAR(128) NOT NULL;
ALTER TABLE slk_user ADD UNIQUE INDEX unique_nickname (nickname);