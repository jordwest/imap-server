DROP DATABASE IF EXISTS imapserver;
CREATE DATABASE imapserver;

USE imapserver;

CREATE TABLE mail_messages (
	`id` BIGINT UNSIGNED PRIMARY KEY auto_increment,
	`uid` INT UNSIGNED,
	`mailbox_id` BIGINT UNSIGNED NOT NULL,
	`date` DATE NOT NULL,
	`flags` SMALLINT UNSIGNED NOT NULL DEFAULT 0,
	`custom_flags` VARCHAR(1024),
	`subject` VARCHAR(255),
	`header` TEXT,
	`body` TEXT
);

CREATE TABLE mail_recipients (
	`id` BIGINT UNSIGNED PRIMARY KEY auto_increment,
	`message_id` BIGINT UNSIGNED NOT NULL,
	`type` ENUM('from', 'to', 'cc', 'bcc') NOT NULL,
	`email_address` VARCHAR(255) NOT NULL,
	`display_name` VARCHAR(255)
);

CREATE TABLE mail_mailboxes (
	`id` BIGINT UNSIGNED PRIMARY KEY auto_increment,
	`name` VARCHAR(50),
	`next_uid` INT UNSIGNED NOT NULL DEFAULT 1
);

INSERT INTO mail_mailboxes (name) VALUES ('INBOX');
INSERT INTO mail_messages (mailbox_id, flags, subject, date) VALUES (LAST_INSERT_ID(), 0, 'Test email', NOW());
INSERT INTO mail_recipients (message_id, type, email_address) VALUES (LAST_INSERT_ID(), 'from', 'test@test.com');
