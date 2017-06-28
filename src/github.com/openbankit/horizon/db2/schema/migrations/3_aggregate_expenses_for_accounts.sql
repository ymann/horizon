-- +migrate Up

CREATE TABLE account_statistics (
    address             character varying(64) NOT NULL,
    asset_code          character varying(12) NOT NULL,
    counterparty_type   smallint NOT NULL DEFAULT 0,
    daily_income 		bigint NOT NULL DEFAULT 0,
    daily_outcome 		bigint NOT NULL DEFAULT 0,
    weekly_income 		bigint	NOT NULL DEFAULT 0,
    weekly_outcome		bigint NOT NULL DEFAULT 0,
    monthly_income		bigint NOT NULL DEFAULT 0,
    monthly_outcome 	bigint NOT NULL DEFAULT 0,
    annual_income		bigint NOT NULL DEFAULT 0,
    annual_outcome		bigint NOT NULL DEFAULT 0,
    updated_at 		    timestamp with time zone NOT NULL,
    PRIMARY KEY(address, asset_code, counterparty_type)
);

CREATE INDEX account_statistics_address_idx ON account_statistics (address);

-- +migrate Down

DROP INDEX account_statistics_address_idx;
DROP TABLE account_statistics;
