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
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'Primary Key ID',
  `from_user_id` bigint unsigned NOT NULL COMMENT 'Sender User ID',
  `to_user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Receiver User ID (0 if group message)',
  `group_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Group ID (0 if private message)',
  `msg_type` tinyint unsigned NOT NULL COMMENT 'Message Type (1:text, 2:image, etc.)',
  `content` text NOT NULL COMMENT 'Message Content',
  `status` tinyint unsigned NOT NULL DEFAULT 1 COMMENT 'Message Status (1:sent, 2:delivered, 3:read)',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Creation Time (Milliseconds)',
  `updated_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Update Time (Milliseconds)',
  PRIMARY KEY (`id`),
  INDEX `idx_from_user` (`from_user_id`),
  INDEX `idx_to_user` (`to_user_id`),
  INDEX `idx_group` (`group_id`),
  INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'Message Table';

CREATE TABLE IF NOT EXISTS `conversation` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'Primary Key ID',
  `user_id` bigint unsigned NOT NULL COMMENT 'User ID',
  `chat_id` bigint unsigned NOT NULL COMMENT 'Chat ID (user_id for private chat, group_id for group chat)',
  `chat_type` tinyint unsigned NOT NULL COMMENT 'Chat Type (1:private, 2:group)',
  `last_msg_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Last Message ID',
  `unread_count` int unsigned NOT NULL DEFAULT 0 COMMENT 'Unread Message Count',
  `updated_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Update Time (Milliseconds)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_chat` (`user_id`, `chat_id`, `chat_type`)
) ENGINE=InnoDB CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'Conversation Table';
