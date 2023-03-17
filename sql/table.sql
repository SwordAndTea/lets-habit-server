CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'record id',
    `uid` varchar(32) NOT NULL COMMENT 'user id',
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

CREATE TABLE IF NOT EXISTS `habits` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key id',
    `name` varchar(255) NOT NULL COMMENT 'habit name',
    `identity_to_form` varchar(255) COMMENT 'identity to form',
    `owner` varchar(32) NOT NULL COMMENT 'habit owner uid',
    `create_at` datetime NOT NULL COMMENT 'create utc time',
    `log_days` tinyint unsigned COMMENT 'days in week need to log, bit mask',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='habit info table';

CREATE TABLE IF NOT EXISTS `habit_groups` (
   `habit_id` bigint unsigned NOT NULL COMMENT 'habit primary key id',
   `uid` varchar(32) NOT NULL COMMENT 'user id',
    PRIMARY KEY (`habit_id`, `uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='user habit join relation';

CREATE TABLE IF NOT EXISTS `user_habit_configs` (
    `uid` varchar(32) NOT NULL COMMENT 'user id',
    `habit_id` bigint unsigned NOT NULL COMMENT 'habit primary key id',
    `current_streak` int unsigned NOT NULL COMMENT 'current consecutive record days',
    `longest_streak` int unsigned NOT NULL COMMENT 'longest consecutive record days',
    `streak_update_at` datetime COMMENT 'when streak info was last updated',
    `remain_retroactive_chance` tinyint unsigned NOT NULL COMMENT 'remain retroactive change',
    `heatmap_color` varchar(8) NOT NULL COMMENT 'heatmap hex rgb color',
    PRIMARY KEY (`habit_id`, `uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='user habit config info';

CREATE TABLE IF NOT EXISTS `habit_log_records` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key id',
    `habit_id` bigint unsigned NOT NULL COMMENT 'habit primary key id',
    `uid` varchar(32) NOT NULL COMMENT 'user id',
    `log_at` datetime NOT NULL COMMENT 'log time',
    PRIMARY KEY (`id`),
    index idx_habit_id(`habit_id`),
    index idx_uid(`uid`),
    index idx_log_time(`log_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='user habit log record';

CREATE TABLE IF NOT EXISTS `unconfirmed_habit_log_records` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key id',
    `habit_id` bigint unsigned NOT NULL COMMENT 'habit primary key id',
    `uid` varchar(32) NOT NULL COMMENT 'user id',
    `log_at` datetime NOT NULL COMMENT 'log time',
    PRIMARY KEY (`id`),
    index idx_habit_id(`habit_id`),
    index idx_uid(`uid`),
    index idx_log_time(`log_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='user habit temporary log record';