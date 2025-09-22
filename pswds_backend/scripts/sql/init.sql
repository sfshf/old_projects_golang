CREATE DATABASE IF NOT EXISTS `pswds` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE pswds;

DROP TABLE IF EXISTS `backup`;
CREATE TABLE `backup` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `created_at` BIGINT NOT NULL COMMENT 'created at',
  `updated_at` BIGINT NOT NULL COMMENT 'updated at',
  `user_id` BIGINT NOT NULL COMMENT 'slark user id',
  `password_hash` VARCHAR(64) NOT NULL COMMENT 'unlock password hash',
  `user_public_key` TEXT NOT NULL COMMENT 'user public key',
  `encrypted_family_key` TEXT NOT NULL COMMENT 'encrypted family key',
  `security_questions` TEXT NOT NULL COMMENT 'unlock password security questions',
  `security_questions_ciphertext` TEXT NOT NULL COMMENT 'unlock password security question answers',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user_id` (user_id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='backup table';

DROP TABLE IF EXISTS `password_record`;
CREATE TABLE `password_record` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `data_id` VARCHAR(64) NOT NULL COMMENT 'uuid of content',
  `updated_at` BIGINT NOT NULL COMMENT 'updated at',
  `user_id` BIGINT NOT NULL COMMENT 'slark user id',
  `content` TEXT NOT NULL COMMENT 'password record, json string',
  `version` BIGINT NOT NULL COMMENT 'record version, extensional field',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_data` (data_id),
  INDEX `pswd_userid` (user_id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='password record table';

DROP TABLE IF EXISTS `trusted_contact`;
CREATE TABLE `trusted_contact` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `created_at` BIGINT NOT NULL COMMENT 'created at',
  `updated_at` BIGINT NOT NULL COMMENT 'updated at',
  `user_id` BIGINT NOT NULL COMMENT 'slark user id',
  `contact_email` VARCHAR(128) NOT NULL COMMENT 'contact email',
  `backup_ciphertext`  VARCHAR(256) NOT NULL COMMENT 'backup ciphertext',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user_id_contact_email` (user_id, contact_email)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='trusted contact table';

DROP TABLE IF EXISTS `privacy_email`;
CREATE TABLE `privacy_email` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `email_account` VARCHAR(512) NOT NULL COMMENT 'email account',
  `mailbox`  VARCHAR(16) NOT NULL COMMENT 'mail box',
  `uid` BIGINT NOT NULL COMMENT 'email uid',
  `sent_at` BIGINT NOT NULL COMMENT 'sent at',
  `sent_by`  VARCHAR(256) NOT NULL COMMENT 'sent by',
  `subject`  VARCHAR(256) NOT NULL COMMENT 'email subject',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='privacy email table';

DROP TABLE IF EXISTS `privacy_email_content`;
CREATE TABLE `privacy_email_content` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `email_id` BIGINT NOT NULL COMMENT 'email id',
  `type`  VARCHAR(16) NOT NULL COMMENT 'type: body|attachment',
  `content_type`  VARCHAR(16) NOT NULL COMMENT 'content type: text|image|bytes',
  `filename`  VARCHAR(256) NOT NULL COMMENT 'file name, if it is an attachment',
  `content` MEDIUMTEXT NOT NULL COMMENT 'email content, base64 string',
  `filesize` BIGINT NOT NULL COMMENT 'file size in bytes, if it is an attachment',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='privacy email content table';

DROP TABLE IF EXISTS `privacy_email_account`;
CREATE TABLE `privacy_email_account` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `user_id` BIGINT NOT NULL COMMENT 'slark user id',
  `email_account` VARCHAR(512) NOT NULL COMMENT 'email account',
  `password`  VARCHAR(512) NOT NULL COMMENT 'email account password',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_email_account` (email_account)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='email account table';

DROP TABLE IF EXISTS `non_password_record`;
CREATE TABLE `non_password_record` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `data_id` VARCHAR(64) NOT NULL COMMENT 'uuid of content',
  `updated_at` BIGINT NOT NULL COMMENT 'updated at',
  `user_id` BIGINT NOT NULL COMMENT 'slark user id',
  `type` VARCHAR(32) NOT NULL COMMENT 'record type',
  `content` TEXT NOT NULL COMMENT 'password record, json string',
  `version` BIGINT NOT NULL COMMENT 'record version, extensional field',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_data` (data_id),
  INDEX `pswd_userid` (user_id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='non password record table';

DROP TABLE IF EXISTS `family`;
CREATE TABLE `family` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `created_at` BIGINT NOT NULL COMMENT 'created at',
  `created_by` BIGINT NOT NULL COMMENT 'slark user id',
  `family_id` VARCHAR(64) NOT NULL COMMENT 'family id',
  `description` VARCHAR(20) NOT NULL COMMENT 'description',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_created_by` (created_by)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='family table';

DROP TABLE IF EXISTS `family_member`;
CREATE TABLE `family_member` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `created_at` BIGINT NOT NULL COMMENT 'created at',
  `user_id` BIGINT NOT NULL COMMENT 'slark user id',
  `family_id` VARCHAR(64) NOT NULL COMMENT 'family id',
  `is_admin` TINYINT NOT NULL COMMENT 'is admin',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user_id` (user_id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='family member table';

DROP TABLE IF EXISTS `family_shared_record`;
CREATE TABLE `family_shared_record` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `data_id` VARCHAR(64) NOT NULL COMMENT 'uuid of content',
  `updated_at` BIGINT NOT NULL COMMENT 'updated at',
  `family_id` VARCHAR(64) NOT NULL COMMENT 'family id',
  `shared_by` BIGINT NOT NULL COMMENT 'slark user id',
  `type` VARCHAR(32) NOT NULL COMMENT 'record type',
  `content` TEXT NOT NULL COMMENT 'password record, json string',
  `shared_to_all` TINYINT NOT NULL COMMENT 'shared to all',
  `sharing_members` TEXT NOT NULL COMMENT 'sharing members, json array string',
  `version` BIGINT NOT NULL COMMENT 'record version, extensional field',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_data` (data_id, shared_by)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='family shared record table';

DROP TABLE IF EXISTS `family_invitation`;
CREATE TABLE `family_invitation` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `created_at` BIGINT NOT NULL COMMENT 'created at',
  `family_id` VARCHAR(64) NOT NULL COMMENT 'family id',
  `invited_by` BIGINT NOT NULL COMMENT 'slark user id',
  `email` VARCHAR(256) NOT NULL COMMENT 'slark email owned by user who has been invited',
  `encrypted_family_key` TEXT NOT NULL COMMENT 'encrypted family key',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='family invitation table';

DROP TABLE IF EXISTS `family_message`;
CREATE TABLE `family_message` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `family_id` VARCHAR(64) NOT NULL COMMENT 'family id',
  `created_at` BIGINT NOT NULL COMMENT 'created at',
  `created_by` VARCHAR(256) NOT NULL COMMENT 'slark user email',
  `target` VARCHAR(256) NOT NULL COMMENT 'slark user email',
  `operation` TINYINT NOT NULL COMMENT 'operation',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='family message table';

DROP TABLE IF EXISTS `family_backup`;
CREATE TABLE `family_backup` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `family_id` VARCHAR(64) NOT NULL COMMENT 'family id',
  `created_at` BIGINT NOT NULL COMMENT 'created at',
  `created_by` BIGINT NOT NULL COMMENT 'slark user id',
  `member_id` BIGINT NOT NULL COMMENT 'slark user id',
  `member` VARCHAR(256) NOT NULL COMMENT 'slark user email',
  `ciphertext` TEXT NOT NULL COMMENT 'ciphertext',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='family backup table';

DROP TABLE IF EXISTS `family_recover`;
CREATE TABLE `family_recover` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `uuid` VARCHAR(64) NOT NULL COMMENT 'uuid',
  `backup_id` BIGINT NOT NULL COMMENT 'backup id',
  `created_at` BIGINT NOT NULL COMMENT 'created at',
  `checked_at` BIGINT NOT NULL COMMENT 'checked at: reject or approve',
  `created_by` BIGINT NOT NULL COMMENT 'slark user id',
  `target_id` BIGINT NOT NULL COMMENT 'slark user id',
  `operation` TINYINT NOT NULL COMMENT 'operation',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='family recover table';

DROP TABLE IF EXISTS `unlock_password_recover`;
CREATE TABLE `unlock_password_recover` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `created_at` BIGINT NOT NULL COMMENT 'created at',
  `created_by` BIGINT NOT NULL COMMENT 'slark user id',
  `type` TINYINT NOT NULL COMMENT 'operation type',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='unlock password recover table';