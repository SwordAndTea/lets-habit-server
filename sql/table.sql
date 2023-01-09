CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'record id',
    `uid` varchar(36) NOT NULL COMMENT 'user id',
    `name` varchar(32) COMMENT 'user name',
    `email` varchar(32) COMMENT 'email',
    `email_active` bool COMMENT 'flag indicate whether user email activated',
    `email_bind` bool COMMENT 'flag indicate whether user email bound',
    `password` varchar(64) COMMENT 'password',
    `portrait` varchar(64) COMMENT 'portrait object storage key',
    `user_register_type` varchar(16) NOT NULL COMMENT 'user register type',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_uid` (`uid`),
    UNIQUE KEY `uniq_name` (`name`),
    UNIQUE KEY `uniq_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='all registered users table';
