-- +migrate Up

CREATE TABLE options
(
  name varchar(32) NOT NULL,
  data text,
  PRIMARY KEY(name)
);


-- +migrate Down

DROP TABLE options;