# CS:GO Hub backend

## About
This is the backend, which works with Steam's OpenID to gather their Steam ID and convert it to the corresponding CS:GO friend code.
It functions with the [CS:GO Hub Discord bot](https://github.com/jesperbakhandskemager/csgo-hub-discord).

A user can request a token from the bot by issuing the `/link-steam` command in any server where the bot is present or in the bot's DM's.
The bot takes note of the Discord Id, and sends it along in the token request (the endpoint can only be accessed from localhost for security).

The bot responds to the user with a formatted link to the `steam.csgohub.xyz` site appended by their token, and the backend gathers various informations from both the database along with data from Discord's API.

Once a user is linked, anyone can issue the `/show-team` command in any channel the server owner permits and the bot will reply with a list of friend codes for any linked users in the same voice channel.

## Database
You need the following tables in MySQL

```{sql}
CREATE TABLE
  `tokens` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
    `discord_id` varchar(21) COLLATE utf8mb4_unicode_ci NOT NULL,
    `token` varchar(8) COLLATE utf8mb4_unicode_ci NOT NULL,
    PRIMARY KEY (`id`)
  ) ENGINE = InnoDB AUTO_INCREMENT = 17 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci
CREATE TABLE
  `users` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
    `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
    `discord_id` varchar(21) COLLATE utf8mb4_unicode_ci NOT NULL,
    `friend_code` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT 'NULL',
    PRIMARY KEY (`id`)
  ) ENGINE = InnoDB AUTO_INCREMENT = 6 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci
```