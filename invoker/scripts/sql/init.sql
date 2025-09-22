CREATE DATABASE IF NOT EXISTS `invoker` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE invoker;

DROP TABLE IF EXISTS `site`;
CREATE TABLE `site` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id', 
  `name` VARCHAR(32) NOT NULL COMMENT 'name',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_name` (name)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='site table';

DROP TABLE IF EXISTS `site_admin`;
CREATE TABLE `site_admin` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `site_id` BIGINT NOT NULL COMMENT 'site id',
  `user_id` BIGINT NOT NULL COMMENT 'admin user id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_site_admin` (site_id, user_id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='site admin table';

DROP TABLE IF EXISTS `category`;
CREATE TABLE `category` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id', 
  `site_id` BIGINT NOT NULL COMMENT 'site id',
  `name` VARCHAR(32) NOT NULL COMMENT 'name',
  `posts` BIGINT NOT NULL COMMENT 'posts count',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_site_category` (site_id, name)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='category table';

DROP TABLE IF EXISTS `post`;
CREATE TABLE `post` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `site_id` BIGINT NOT NULL COMMENT 'site id',
  `category_id` BIGINT NOT NULL COMMENT 'category id',
  `title` VARCHAR(256) NOT NULL COMMENT 'title',
  `posted_at` BIGINT NOT NULL COMMENT 'posted at',
  `posted_by` BIGINT NOT NULL COMMENT 'posted by',
  `content` TEXT NOT NULL COMMENT 'content',
  `image` VARCHAR(256) NULL COMMENT 'image url',
  `state` TINYINT NOT NULL COMMENT 'state',
  `views` BIGINT NOT NULL COMMENT 'views',
  `replies` BIGINT NOT NULL COMMENT 'replies',
  `thumbups` BIGINT NOT NULL COMMENT 'thumbups',
  `activity` BIGINT NOT NULL COMMENT 'activity',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='post table';

DROP TABLE IF EXISTS `comment`;
CREATE TABLE `comment` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `site_id` BIGINT NOT NULL COMMENT 'site id',
  `category_id` BIGINT NOT NULL COMMENT 'category id',
  `post_id` BIGINT NOT NULL COMMENT 'post id',
  `root_comment_id` BIGINT NOT NULL COMMENT 'root comment id',
  `content` TEXT NOT NULL COMMENT 'content',
  `posted_at` BIGINT NOT NULL COMMENT 'posted at',
  `posted_by` BIGINT NULL COMMENT 'posted by',
  `at_who` BIGINT NOT NULL COMMENT 'at who, user id',
  `updated_at` BIGINT NOT NULL COMMENT 'updated at',
  `replies` BIGINT NOT NULL COMMENT 'replies',
  `thumbups` BIGINT NOT NULL COMMENT 'thumbups',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='comment table';

DROP TABLE IF EXISTS `thumbup`;
CREATE TABLE `thumbup` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'primary id',
  `site_id` BIGINT NOT NULL COMMENT 'site id',
  `category_id` BIGINT NOT NULL COMMENT 'category id',
  `post_id` BIGINT NOT NULL COMMENT 'post id',
  `comment_id` BIGINT NULL COMMENT 'comment id',
  `type` TINYINT NOT NULL COMMENT 'type: post/comment',
  `posted_at` BIGINT NOT NULL COMMENT 'posted at',
  `posted_by` BIGINT NULL COMMENT 'posted by',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='thumbup table';