CREATE DATABASE IF NOT EXISTS `connector` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE connector;

DROP TABLE IF EXISTS `app_config`;
CREATE TABLE `app_config` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `app` VARCHAR(256) NOT NULL,
                        `config` TEXT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`),
                        UNIQUE KEY(app,deleted_at) 
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `relation_app_key`;
CREATE TABLE `relation_app_key` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `app` VARCHAR(256) NOT NULL,
                        `key_id` VARCHAR(256) NOT NULL,
                        `password_hash` VARCHAR(256) NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `relation_app_data`;
CREATE TABLE `relation_app_data` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `app` VARCHAR(256) NOT NULL,
                        `key_id` VARCHAR(256) NOT NULL,
                        `data_id` VARCHAR(256) NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `api_key`;
CREATE TABLE `api_key` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `app` VARCHAR(256) NOT NULL,
                        `key_id` VARCHAR(256) NOT NULL,
                        `name` VARCHAR(256) NOT NULL,
                        `permission` VARCHAR(8) NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `manage_platform_log`;
CREATE TABLE `manage_platform_log` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `ip` VARCHAR(256) NOT NULL,
                        `api_key_name` VARCHAR(256) NOT NULL,
                        `status` VARCHAR(8) NOT NULL,
                        `object` VARCHAR(256) NULL,
                        `operation` VARCHAR(256) NOT NULL,
                        `key_id` VARCHAR(256) NULL,
                        `app` VARCHAR(256) NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;