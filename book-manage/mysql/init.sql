CREATE DATABASE IF NOT EXISTS `word` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `word`;

DROP TABLE IF EXISTS `book`;
CREATE TABLE `book` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `name` VARCHAR(128) NOT NULL,
                        `description` TEXT NOT NULL,
                        `download_url` VARCHAR(512) NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `related_book`;
CREATE TABLE `related_book` (
                                `id` BIGINT NOT NULL AUTO_INCREMENT,
                                `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                                `item_id` BIGINT NOT NULL,
                                `item_type` VARCHAR(128) NOT NULL,
                                `book_id` BIGINT NOT NULL,
                                `sort_value` INT NOT NULL,
                                INDEX `index_deleted` (deleted_at),
                                INDEX `index_book` (book_id),
                                PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;


DROP TABLE IF EXISTS `definition`;
CREATE TABLE `definition` (
                              `id` BIGINT NOT NULL AUTO_INCREMENT,
                              `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                              `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                              `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',

                              `string_id` BIGINT NOT NULL,
                              `part_of_speech` VARCHAR(128) NOT NULL,
                              `specific_type` VARCHAR(128),
                              `pronunciation_ipa` VARCHAR(128),
                              `pronunciation_ipa_weak` VARCHAR(128),
                              `pronunciation_ipa_other` VARCHAR(128),
                              `pronunciation_text` VARCHAR(512),
    -- `pronunciation_ssml` VARCHAR(512),
                              `cefr_level` VARCHAR(16),
                              `definition` TEXT NOT NULL,
                              PRIMARY KEY (`id`),
                              INDEX `index_deleted` (deleted_at),
                              INDEX `index_string_id` (string_id)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;


DROP TABLE IF EXISTS `related_definition`;
CREATE TABLE `related_definition` (
                                      `id` BIGINT NOT NULL AUTO_INCREMENT,
                                      `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                      `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                      `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                                      `definition_id` BIGINT NOT NULL,
                                      `related_definition_id` BIGINT NOT NULL,
                                      PRIMARY KEY (`id`),
                                      INDEX `index_deleted` (deleted_at),
                                      INDEX `index_definition_id` (definition_id)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;


DROP TABLE IF EXISTS `example`;
CREATE TABLE `example` (
                           `id` BIGINT NOT NULL AUTO_INCREMENT,
                           `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                           `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                           `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                           `string_id` BIGINT NOT NULL,
                           `definition_id` BIGINT NOT NULL,
                           `content` TEXT NOT NULL,
                           `word_positions` VARCHAR(256) NOT NULL COMMENT 'index1,length1,index2,length2',
                           `favour_count` INT NOT NULL DEFAULT 0,
                           PRIMARY KEY (`id`),
                           INDEX `index_deleted` (deleted_at),
                           INDEX `index_string_id` (string_id),
                           INDEX `index_definition` (definition_id)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;


DROP TABLE IF EXISTS `string`;
CREATE TABLE `string` (
                          `id` BIGINT NOT NULL AUTO_INCREMENT,
                          `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                          `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                          `string` VARCHAR(256) NOT NULL,
                          `type` VARCHAR(64) NOT NULL,
                          `base_string_id` BIGINT,
                          PRIMARY KEY (`id`),
                          INDEX `index_deleted` (deleted_at),
                          INDEX `index_string` (string),
                          INDEX `index_base_id` (base_string_id)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `operate_log`;
CREATE TABLE `operate_log` (
                          `id` BIGINT NOT NULL AUTO_INCREMENT,
                          `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'operate time',
                          `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                          `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                          `created_by` VARCHAR(256) NOT NULL COMMENT 'operator',
                          `operate_status` TINYINT NOT NULL COMMENT '1:failure/2:success',
                          `operate_type` VARCHAR(64) NOT NULL COMMENT 'operate type/api',
                          `book_id` BIGINT NOT NULL COMMENT 'bookID',
                          `definition_id` BIGINT NULL COMMENT 'definitionID',
                          `other_operate_params` JSON NULL COMMENT 'other operation params',
                          `error` VARCHAR(512) DEFAULT NULL COMMENT 'error message',
                          PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;


DROP TABLE IF EXISTS `definition_comment`;
CREATE TABLE `definition_comment` (
                          `id` BIGINT NOT NULL AUTO_INCREMENT,
                          `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                          `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                          `definition_id` BIGINT NOT NULL COMMENT 'definitionID',
                          `content` TEXT NOT NULL,
                          PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE `definition_comment` MODIFY `definition_id` BIGINT NOT NULL COMMENT 'definition id';


DROP TABLE IF EXISTS `translation`;
CREATE TABLE `translation` (
                          `id` BIGINT NOT NULL AUTO_INCREMENT,
                          `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                          `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                          `item_type` VARCHAR(128) NOT NULL COMMENT 'definition or example',
                          `item_id` BIGINT NOT NULL COMMENT 'definition id or example id',
                          `content` TEXT NOT NULL COMMENT 'content of the translation',
                          `language_code` VARCHAR(4) NOT NULL COMMENT 'ISO 639-1 language code',
                          PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;


DROP TABLE IF EXISTS `backup`;
CREATE TABLE `backup` (
                          `id` BIGINT NOT NULL AUTO_INCREMENT,
                          `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'the time when make a backup',
                          `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                          `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                          `book_id` BIGINT NOT NULL,
                          `filepath` VARCHAR(256) NOT NULL COMMENT 'path with name of the backup file',
                          PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;
