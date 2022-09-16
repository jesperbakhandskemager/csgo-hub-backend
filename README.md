# CS:GO Hub backend
Use with [csgo-hub-discord](https://github.com/jesperbakhandskemager/csgo-hub-discord)

## Database
You need the following tables in MySQL

```{sql}
CREATE TABLE
  `tokens` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
    `discord_id` varchar(18) COLLATE utf8mb4_unicode_ci NOT NULL,
    `token` varchar(8) COLLATE utf8mb4_unicode_ci NOT NULL,
    PRIMARY KEY (`id`)
  ) ENGINE = InnoDB AUTO_INCREMENT = 17 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci
CREATE TABLE
  `users` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
    `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
    `discord_id` varchar(18) COLLATE utf8mb4_unicode_ci NOT NULL,
    `friend_code` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT 'NULL',
    PRIMARY KEY (`id`)
  ) ENGINE = InnoDB AUTO_INCREMENT = 6 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci
```