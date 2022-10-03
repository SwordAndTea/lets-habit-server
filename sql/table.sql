CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'record id',
    `uid` varchar(36) NOT NULL COMMENT 'user id',
    `name` varchar(32) COMMENT 'user name',
    `email` varchar(32) COMMENT 'email',
    `password` varchar(64) COMMENT 'password',
    `portrait` varchar(64) COMMENT 'portrait object storage key',
    `user_register_type` varchar(16) NOT NULL COMMENT 'user register type',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_uid` (`uid`),
    UNIQUE KEY `uniq_name` (`name`),
    UNIQUE KEY `uniq_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='all registered users table';


CREATE TABLE IF NOT EXISTS `user_email_activates` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'record id',
    `uid` varchar(36) NOT NULL COMMENT 'user id',
    `email` varchar(32) NOT NULL COMMENT 'email',
    `password` varchar(64) NOT NULL COMMENT 'password',
    `send_at` datetime NOT NULL COMMENT 'activate code send time',
    `activated` bool NOT NULL COMMENT 'activated or not',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_uid` (`uid`),
    UNIQUE KEY `uniq_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='user email register info table';