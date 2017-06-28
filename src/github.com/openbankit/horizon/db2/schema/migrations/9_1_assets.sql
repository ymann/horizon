-- +migrate Up

CREATE TABLE asset (
    id           bigserial,
	type         int NOT NULL,
    code         character varying(12) NOT NULL,
    issuer       character varying(64) NOT NULL,
    is_anonymous boolean NOT NULL,
    PRIMARY KEY(id)
);

CREATE UNIQUE INDEX assets_code_issuer_type ON asset (code, issuer, type);

-- +migrate Down

DROP TABLE asset;