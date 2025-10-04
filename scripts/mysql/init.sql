SET NAMES utf8mb4;
CREATE DATABASE IF NOT EXISTS goim COLLATE utf8mb4_unicode_ci;

USE goim;

CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'Primary Key ID',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT 'User Nickname',
  `unique_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'User Unique Name',
  `email` varchar(128) NOT NULL DEFAULT '' COMMENT 'Email',
  `password` varchar(128) NOT NULL DEFAULT '' COMMENT 'Password (Encrypted)',
  `description` varchar(512) NOT NULL DEFAULT '' COMMENT 'User Description',
  `icon_uri` varchar(512) NOT NULL DEFAULT '' COMMENT 'Avatar URI',
  `sex` tinyint DEFAULT NULL COMMENT 'User sex',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Creation Time (Milliseconds)',
  `updated_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Update Time (Milliseconds)',
  `deleted_at` bigint unsigned NULL COMMENT 'Deletion Time (Milliseconds)',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uniq_email` (`email`),
  UNIQUE INDEX `uniq_unique_name` (`unique_name`)
) ENGINE=InnoDB CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'User Table';

CREATE TABLE IF NOT EXISTS `message` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'Message ID',
  `send_id` bigint NOT NULL COMMENT 'Sender ID',
  `recv_id` bigint NOT NULL COMMENT 'Receiver ID',
  `group_id` bigint NOT NULL COMMENT 'Group ID',
  `client_msg_id` bigint NOT NULL COMMENT 'Client Message ID',
  `session_type` int NOT NULL COMMENT 'Session Type',
  `message_from` int NOT NULL COMMENT 'Message Source',
  `content_type` int NOT NULL COMMENT 'Message Content Type',
  `content` text NOT NULL COMMENT 'Message Content',
  `seq` bigint NOT NULL COMMENT 'Message Sequence Number',
  `send_time` bigint NOT NULL COMMENT 'Send Time (Milliseconds)',
  `status` int NOT NULL COMMENT 'Message Status',
  `is_read` boolean NOT NULL COMMENT 'Message Read Status',
  `created_time` bigint NOT NULL COMMENT 'Creation Time (Milliseconds)',
  `updated_time` bigint NOT NULL COMMENT 'Update Time (Milliseconds)',
  PRIMARY KEY (`id`),
  INDEX `idx_send_id` (`send_id`),
  INDEX `idx_recv_id` (`recv_id`),
  INDEX `idx_group_id` (`group_id`)
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT 'Message Table';