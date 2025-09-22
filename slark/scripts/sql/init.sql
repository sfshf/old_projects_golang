CREATE DATABASE IF NOT EXISTS `slark` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE slark;

DROP TABLE IF EXISTS `slk_device_extra`;
CREATE TABLE `slk_device_extra` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键id', 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` BIGINT NOT NULL COMMENT '删除标记位',
  `application` VARCHAR(32) NOT NULL COMMENT '应用',
  `did` BIGINT NOT NULL COMMENT 'device表id 外键',
  `key` VARCHAR(64) NOT NULL COMMENT 'key',
  `value` TEXT NOT NULL COMMENT 'value',
  PRIMARY KEY (`id`),
  INDEX `index_did` (did, application(8), deleted_at),
  UNIQUE KEY `unique_key` (did, `key` , application, deleted_at),
  INDEX `index_id` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='设备额外数据表';


DROP TABLE IF EXISTS `slk_device`;
CREATE TABLE `slk_device` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键id', 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` BIGINT NOT NULL COMMENT '删除标记位',
  `device_id` VARCHAR(32) UNIQUE NOT NULL COMMENT '设备ID',
  `platform` VARCHAR(16) NOT NULL COMMENT '系统平台',
  `device_model` VARCHAR(64) NOT NULL COMMENT '设备型号',
  `resolution_width` INT NOT NULL COMMENT '分辨率 宽，px',
  `resolution_height` INT NOT NULL COMMENT '分辨率 高，px',
  `screen_density` FLOAT NOT NULL COMMENT '屏幕密度',
  `rom` FLOAT NOT NULL COMMENT '内存,单位gb',
  `ram` FLOAT NOT NULL COMMENT '磁盘，单位gb',
  PRIMARY KEY (`id`),
  INDEX `index_device_id` (device_id, deleted_at),
  INDEX `index_id` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='设备信息表';


DROP TABLE IF EXISTS `slk_session`;
CREATE TABLE `slk_session` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键id', 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` BIGINT NOT NULL COMMENT '删除标记位',
  `application` VARCHAR(32) NOT NULL COMMENT '应用',
  `user_id` BIGINT NOT NULL COMMENT '用户id， 外键',
  `session_id` VARCHAR(32) NOT NULL COMMENT 'session ID',
  `device_id` VARCHAR(32) NOT NULL COMMENT '设备id， 非设备表ID',
  `login_ip` VARCHAR(40) NOT NULL COMMENT 'ip, v4或v6',
  PRIMARY KEY (`id`),
  INDEX `index_1session_1app` (user_id, device_id(16), application(8), deleted_at),
  INDEX `index_user_id` (user_id, application(8), deleted_at),
  INDEX `index_session_id` (session_id(16), application(8), deleted_at),
  INDEX `index_id` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='session 表';


DROP TABLE IF EXISTS `slk_user_extra`;
CREATE TABLE `slk_user_extra` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键id', 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` BIGINT NOT NULL COMMENT '删除标记位',
  `application` VARCHAR(32) NOT NULL COMMENT '应用',
  `user_id` BIGINT NOT NULL COMMENT '用户id， 外键',
  `key` VARCHAR(64) NOT NULL COMMENT 'key',
  `value` TEXT NOT NULL COMMENT 'value',
  PRIMARY KEY (`id`),
  INDEX `index_user_id` (user_id, application(8), deleted_at),
  UNIQUE KEY `unique_key` (user_id, `key`, application, deleted_at),
  INDEX `index_id` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='用户额外数据表';


DROP TABLE IF EXISTS `slk_third_party`;
CREATE TABLE `slk_third_party` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键id', 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` BIGINT NOT NULL COMMENT '删除标记位',
  `application` VARCHAR(32) NOT NULL COMMENT '应用， 第三方登录是针对单个APP的',
  `user_id` BIGINT NOT NULL COMMENT '用户id',
  `union_id` VARCHAR(64) NOT NULL COMMENT '第三方唯一id',
  `open_id` VARCHAR(64) NOT NULL COMMENT '第三方id',
  `access_token` VARCHAR(64) NOT NULL COMMENT 'token',
  `third_party` VARCHAR(64) NOT NULL COMMENT '第三方 名称',
  `extra_info` TEXT COMMENT '第三方额外信息',
  PRIMARY KEY (`id`),
  INDEX `index_union_id` (union_id(16), application(8), third_party, deleted_at),
  INDEX `index_open_id` (open_id(16), application(8), third_party, deleted_at),
  INDEX `id_third_app` (id, application(8), third_party, deleted_at),
  INDEX `user_id` (user_id, deleted_at),
  INDEX `index_id` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='第三方登录信息表';


DROP TABLE IF EXISTS `slk_user`;
CREATE TABLE `slk_user` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键id', 
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` BIGINT NOT NULL COMMENT '删除标记位',
  `nickname` VARCHAR(64) NOT NULL COMMENT '用户昵称',
  `password_hash` VARCHAR(64) NOT NULL COMMENT '用户密码HASH值',
  `first_name` VARCHAR(64) COMMENT '用户名',
  `last_name` VARCHAR(64) COMMENT '用户姓',
  `email` VARCHAR(64) NOT NULL  COMMENT 'email, 可为空',
  `phone` VARCHAR(16) NOT NULL COMMENT '手机号, 可为空',
  `gender` TINYINT COMMENT '性别',
  `birthday` DATE NULL COMMENT '出生日期',
  PRIMARY KEY (`id`),
  INDEX `index_email` (email, deleted_at),
  INDEX `index_phone` (phone, deleted_at),
  INDEX `index_id` (id, deleted_at)
) ENGINE=InnoDB AUTO_INCREMENT=100000000 DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

