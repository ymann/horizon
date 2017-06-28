-- +migrate Up

CREATE TABLE account_limits (
	address             character varying(64) NOT NULL,
    asset_code          character varying(12) NOT NULL,   
    max_operation 		bigint NOT NULL DEFAULT 0,
    daily_turnover 		bigint NOT NULL DEFAULT 0,
    monthly_turnover	bigint	NOT NULL DEFAULT 0,
    PRIMARY KEY(address, asset_code)
);

-- +migrate Down

DROP TABLE account_limits;
