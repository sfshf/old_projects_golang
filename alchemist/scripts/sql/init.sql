CREATE DATABASE IF NOT EXISTS `alchemist` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE alchemist;

DROP TABLE IF EXISTS `slark_users`;
CREATE TABLE `slark_users` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `app_account_token` VARCHAR(64) NOT NULL,
                        `registered_at` BIGINT NOT NULL,
                        `user_id` BIGINT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `raw_transactions`;
CREATE TABLE `raw_transactions` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `data` TEXT NOT NULL,
                        `handled` BOOLEAN NOT NULL,
                        `error` TEXT NOT NULL,
                        `environment` TINYINT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `subscription_state_prod`;
CREATE TABLE `subscription_state_prod` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `user_id` BIGINT NOT NULL,
                        `app` VARCHAR(64) NOT NULL,
                        `subscribed` BOOLEAN NOT NULL,
                        `first_time_subscribed_at` BIGINT NOT NULL,
                        `subscribed_at` BIGINT NOT NULL,
                        `current_bill_date` BIGINT NOT NULL,
                        `current_bill_price` VARCHAR(32) NOT NULL,
                        `current_offer_type` VARCHAR(64) NOT NULL,
                        `current_offer_id` VARCHAR(64) NOT NULL,
                        `expiration_date` BIGINT NOT NULL,
                        `auto_renew` BOOLEAN NOT NULL,
                        `next_bill_price` VARCHAR(32) NOT NULL,
                        `next_offer_type` VARCHAR(64) NOT NULL,
                        `next_offer_id` VARCHAR(64) NOT NULL,
                        `currency_code` VARCHAR(32) NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `subscription_state_test`;
CREATE TABLE `subscription_state_test` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `user_id` BIGINT NOT NULL,
                        `app` VARCHAR(64) NOT NULL,
                        `subscribed` BOOLEAN NOT NULL,
                        `first_time_subscribed_at` BIGINT NOT NULL,
                        `subscribed_at` BIGINT NOT NULL,
                        `current_bill_date` BIGINT NOT NULL,
                        `current_bill_price` VARCHAR(32) NOT NULL,
                        `current_offer_type` VARCHAR(64) NOT NULL,
                        `current_offer_id` VARCHAR(64) NOT NULL,
                        `expiration_date` BIGINT NOT NULL,
                        `auto_renew` BOOLEAN NOT NULL,
                        `next_bill_price` VARCHAR(32) NOT NULL,
                        `next_offer_type` VARCHAR(64) NOT NULL,
                        `next_offer_id` VARCHAR(64) NOT NULL,
                        `currency_code` VARCHAR(32) NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `transactions_prod`;
CREATE TABLE `transactions_prod` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `user_id` BIGINT NOT NULL,
                        `app` VARCHAR(64) NOT NULL,
                        `original_transaction_id` VARCHAR(64) NOT NULL,
                        `transaction_id` VARCHAR(64) NOT NULL,
                        `web_order_line_item_id` VARCHAR(64) NOT NULL,
                        `bundle_id` VARCHAR(64) NOT NULL,
                        `app_account_token` VARCHAR(64) NOT NULL,
                        `product_id` VARCHAR(64) NOT NULL,
                        `type` VARCHAR(64) NOT NULL,
                        `subscription_group_identifier` VARCHAR(64) NOT NULL,
                        `quantity` INT NOT NULL,
                        `price` INT NOT NULL,
                        `currency` VARCHAR(32) NOT NULL,
                        `storefront` VARCHAR(64) NOT NULL,
                        `storefront_id` VARCHAR(64) NOT NULL,
                        `offer_identifier` VARCHAR(64) NOT NULL,
                        `offer_type` INT NOT NULL,
                        `offer_discount_type` VARCHAR(64) NOT NULL,
                        `original_purchase_date` BIGINT NOT NULL,
                        `purchase_date` BIGINT NOT NULL,
                        `recent_subscription_start_date` BIGINT NOT NULL,
                        `is_in_billing_retry_period` BOOLEAN NOT NULL,
                        `grace_period_expires_date` BIGINT NOT NULL,
                        `auto_renew_status` INT NOT NULL,
                        `auto_renew_product_id` VARCHAR(64) NOT NULL,
                        `expiration_intent` INT NOT NULL,
                        `expires_date` BIGINT NOT NULL,
                        `is_upgraded` BOOLEAN NOT NULL,
                        `renewal_date` BIGINT NOT NULL,
                        `in_app_ownership_type` VARCHAR(64) NOT NULL,
                        `price_increase_status` INT NOT NULL,
                        `revocation_date` BIGINT NOT NULL,
                        `revocation_reason` INT NOT NULL,
                        `transaction_reason` VARCHAR(528) NOT NULL,
                        `signed_date` BIGINT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `transactions_test`;
CREATE TABLE `transactions_test` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `user_id` BIGINT NOT NULL,
                        `app` VARCHAR(64) NOT NULL,
                        `original_transaction_id` VARCHAR(64) NOT NULL,
                        `transaction_id` VARCHAR(64) NOT NULL,
                        `web_order_line_item_id` VARCHAR(64) NOT NULL,
                        `bundle_id` VARCHAR(64) NOT NULL,
                        `app_account_token` VARCHAR(64) NOT NULL,
                        `product_id` VARCHAR(64) NOT NULL,
                        `type` VARCHAR(64) NOT NULL,
                        `subscription_group_identifier` VARCHAR(64) NOT NULL,
                        `quantity` INT NOT NULL,
                        `price` INT NOT NULL,
                        `currency` VARCHAR(32) NOT NULL,
                        `storefront` VARCHAR(64) NOT NULL,
                        `storefront_id` VARCHAR(64) NOT NULL,
                        `offer_identifier` VARCHAR(64) NOT NULL,
                        `offer_type` INT NOT NULL,
                        `offer_discount_type` VARCHAR(64) NOT NULL,
                        `original_purchase_date` BIGINT NOT NULL,
                        `purchase_date` BIGINT NOT NULL,
                        `recent_subscription_start_date` BIGINT NOT NULL,
                        `is_in_billing_retry_period` BOOLEAN NOT NULL,
                        `grace_period_expires_date` BIGINT NOT NULL,
                        `auto_renew_status` INT NOT NULL,
                        `auto_renew_product_id` VARCHAR(64) NOT NULL,
                        `expiration_intent` INT NOT NULL,
                        `expires_date` BIGINT NOT NULL,
                        `is_upgraded` BOOLEAN NOT NULL,
                        `renewal_date` BIGINT NOT NULL,
                        `in_app_ownership_type` VARCHAR(64) NOT NULL,
                        `price_increase_status` INT NOT NULL,
                        `revocation_date` BIGINT NOT NULL,
                        `revocation_reason` INT NOT NULL,
                        `transaction_reason` VARCHAR(528) NOT NULL,
                        `signed_date` BIGINT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `referral_code`;
CREATE TABLE `referral_code` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `user_id` BIGINT NOT NULL,
                        `app` VARCHAR(64) NOT NULL,
                        `join_date` TIMESTAMP NOT NULL,
                        `referral_code` VARCHAR(16) NOT NULL UNIQUE,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `referral_new_user`;
CREATE TABLE `referral_new_user` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `user_id` BIGINT NOT NULL,
                        `app` VARCHAR(64) NOT NULL,
                        `bind_date` BIGINT NOT NULL,
                        `expired_date` BIGINT NOT NULL,
                        `referral_code` VARCHAR(16) NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `referral_point`;
CREATE TABLE `referral_point` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `user_id` BIGINT NOT NULL,
                        `app` VARCHAR(64) NOT NULL,
                        `points` INT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `referral_log`;
CREATE TABLE `referral_log` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `referral_point_id` BIGINT NOT NULL,
                        `user_id` BIGINT NOT NULL,
                        `app` VARCHAR(64) NOT NULL,
                        `timestamp` BIGINT NOT NULL,
                        `type` TINYINT NOT NULL,
                        `reason` TINYINT NOT NULL,
                        `points` INT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `promo_offer_records`;
CREATE TABLE `promo_offer_records` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `app` VARCHAR(64) NOT NULL,
                        `user_id` BIGINT NOT NULL,
                        `offer_id` VARCHAR(256) NOT NULL,
                        `sign_date` BIGINT NOT NULL,
                        `environment` TINYINT NOT NULL,
                        `use_date` BIGINT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `new_user_discount_state`;
CREATE TABLE `new_user_discount_state` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `app` VARCHAR(64) NOT NULL,
                        `user_id` BIGINT NOT NULL,
                        `referral_code` VARCHAR(256) NOT NULL,
                        `start_date` BIGINT NOT NULL,
                        `billed_times` TINYINT NOT NULL,
                        `remaining_times` TINYINT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;


DROP TABLE IF EXISTS `free_trial_state`;
CREATE TABLE `free_trial_state` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `app` VARCHAR(64) NOT NULL,
                        `user_id` BIGINT NOT NULL,
                        `expiration_date` BIGINT NOT NULL,
                        `start_date` BIGINT NOT NULL,
                        `days_of_trial` INT NOT NULL,
                        `total_days_of_trial` INT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;


DROP TABLE IF EXISTS `user_registered_on_old_device`;
CREATE TABLE `user_registered_on_old_device` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `app` VARCHAR(64) NOT NULL,
                        `user_id` BIGINT NOT NULL,
                        `ip` VARCHAR(256) NOT NULL,
                        `referral_code` VARCHAR(256) NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

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

DROP TABLE IF EXISTS `subscription_count`;
CREATE TABLE `subscription_count` (
                        `id` BIGINT NOT NULL AUTO_INCREMENT,
                        `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        `deleted_at` BIGINT NOT NULL COMMENT 'Coding style',
                        `app` VARCHAR(256) NOT NULL,
                        `time` BIGINT NOT NULL,
                        `count` BIGINT NOT NULL,
                        INDEX `index_deleted` (deleted_at),
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARACTER SET=utf8mb4 COLLATE utf8mb4_unicode_ci;

