-- +migrate Up
CREATE TABLE `users` (
  `user_id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(255) NOT NULL COMMENT 'name',
  `email` varchar(255) NOT NULL COMMENT 'email',
  `salt` varchar(255) NOT NULL COMMENT 'salt',
  `salted` varchar(255) NOT NULL COMMENT 'salted password',
  `created` timestamp NOT NULL DEFAULT NOW() COMMENT 'when created',
  `updated` timestamp NOT NULL DEFAULT NOW() ON UPDATE CURRENT_TIMESTAMP COMMENT 'when last updated',
  PRIMARY KEY (`user_id`),
  UNIQUE KEY (`email`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COMMENT='list of users';

-- +migrate Down
DROP TABLE users;
