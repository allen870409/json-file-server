CREATE DATABASE json-file-server

CREATE TABLE `json_file` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `path` varchar(64) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8