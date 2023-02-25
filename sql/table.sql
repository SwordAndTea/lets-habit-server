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
    `check_deadline_delay` tinyint unsigned COMMENT 'check time delay duration in hours',
    `creator` varchar(32) NOT NULL COMMENT 'habit creator uid',
    `create_at` datetime NOT NULL COMMENT 'create utc time',
    `check_days` tinyint unsigned COMMENT 'days in week need to check, bit mask',
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
    `remain_recheck_chance` tinyint unsigned NOT NULL COMMENT 'remain recheck change',
    `heatmap_color` varchar(8) NOT NULL COMMENT 'heatmap hex rgb color',
    PRIMARY KEY (`habit_id`, `uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='user habit config info';

CREATE TABLE IF NOT EXISTS `habit_check_records` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key id',
    `habit_id` bigint unsigned NOT NULL COMMENT 'habit primary key id',
    `uid` varchar(32) NOT NULL COMMENT 'user id',
    `check_at` datetime NOT NULL COMMENT 'check time',
    PRIMARY KEY (`id`),
    index idx_habit_id(`habit_id`),
    index idx_uid(`uid`),
    index idx_check_time(`check_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='user habit check record';