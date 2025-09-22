CREATE DATABASE IF NOT EXISTS `oracle` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE oracle;

DROP TABLE IF EXISTS `application`;
CREATE TABLE `application` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `unique_name` (name),
  INDEX `index_id_deleted_at` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `service`;
CREATE TABLE `service` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `application_id` BIGINT NOT NULL,
  `url` VARCHAR(256) NOT NULL,
  `path_prefix` VARCHAR(64) NOT NULL,
  `proto_file` TEXT NOT NULL,
  `proto_file_md5` VARCHAR(32) NOT NULL,
  `file_descriptor_data` MEDIUMTEXT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `unique_name` (name),
  UNIQUE `unique_path_prefix` (path_prefix),
  INDEX `index_id_deleted_at` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `acme_resource`;
CREATE TABLE `acme_resource` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT NOT NULL,
  `domain` VARCHAR(256) NOT NULL,
  `cert_url` VARCHAR(256) NOT NULL,
  `cert_stable_url` VARCHAR(256) NOT NULL,
  `private_key` TEXT NOT NULL,
  `certificate` TEXT NOT NULL,
  `issuer_certificate` TEXT NOT NULL,
  `csr` TEXT NOT NULL,
  `token` VARCHAR(128) NULL,
  `key_auth` VARCHAR(512) NULL,
  PRIMARY KEY (`id`),
  UNIQUE `unique_domain` (domain),
  INDEX `index_id_deleted_at` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `host_manage`;
CREATE TABLE `host_manage` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT NOT NULL,
  `domain` VARCHAR(64) NOT NULL,
  `raw_url` VARCHAR(256) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `unique_domain` (domain),
  INDEX `index_id_deleted_at` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `gateway_node`;
CREATE TABLE `gateway_node` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `ipv4` VARCHAR(16) NOT NULL,
  `rpc_port` INT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `unique_name` (name),
  INDEX `index_id_deleted_at` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `proto_statistic`;
CREATE TABLE `proto_statistic` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT NOT NULL,
  `date` VARCHAR(16) NOT NULL,
  `application_id` BIGINT NOT NULL,
  `service_id` BIGINT NOT NULL,
  `path` VARCHAR(256) NOT NULL,
  `hit` BIGINT NOT NULL,
  `success_hit` BIGINT NOT NULL,
  `proxy_success_hit` BIGINT NOT NULL,
  `duration_average` BIGINT NOT NULL,
  `duration_min` BIGINT NOT NULL,
  `duration_max` BIGINT NOT NULL,
  `service_duration_average` BIGINT NOT NULL,
  `service_duration_min` BIGINT NOT NULL,
  `service_duration_max` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `unique_date_path` (date, path),
  INDEX `index_id_deleted_at` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `proto_statistic_hourly`;
CREATE TABLE `proto_statistic_hourly` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT NOT NULL,
  `timestamp` BIGINT NOT NULL,
  `gateway_node` VARCHAR(256) NOT NULL,
  `application` VARCHAR(256) NOT NULL,
  `service` VARCHAR(256) NOT NULL,
  `path` VARCHAR(256) NOT NULL,
  `hit` BIGINT NOT NULL,
  `success_hit` BIGINT NOT NULL,
  `proxy_success_hit` BIGINT NOT NULL,
  `duration_average` BIGINT NOT NULL,
  `duration_min` BIGINT NOT NULL,
  `duration_max` BIGINT NOT NULL,
  `service_duration_average` BIGINT NOT NULL,
  `service_duration_min` BIGINT NOT NULL,
  `service_duration_max` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `unique_timestamp_path` (timestamp, path),
  INDEX `index_id_deleted_at` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `alarm_email`;
CREATE TABLE `alarm_email` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT NOT NULL,
  `address` VARCHAR(256) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `unique_address` (address),
  INDEX `index_id_deleted_at` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `rate_limit_rule`;
CREATE TABLE `rate_limit_rule` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT NOT NULL,
  `type` TINYINT NOT NULL,
  `target` VARCHAR(256) NOT NULL,
  `capacity` BIGINT NOT NULL,
  `enabled` BOOLEAN NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `unique_type_target` (type, target),
  INDEX `index_id_deleted_at` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `timeout_statistic`;
CREATE TABLE `timeout_statistic` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT NOT NULL,
  `date` DATE NOT NULL,
  `application_id` BIGINT NOT NULL,
  `service_id` BIGINT NOT NULL,
  `path` VARCHAR(256) NOT NULL,
  `count` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `unique_date_path` (date, path),
  INDEX `index_id_deleted_at` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;