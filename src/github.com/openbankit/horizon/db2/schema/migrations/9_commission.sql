-- +migrate Up

CREATE TABLE commission (
	id          bigserial,
    key_hash    character(64) NOT NULL,
    key_value 	jsonb NOT NULL,
    flat_fee 	bigint NOT NULL DEFAULT 0,
    percent_fee bigint NOT NULL DEFAULT 0,
    PRIMARY KEY(id)
);

CREATE UNIQUE INDEX commission_by_hash ON commission USING btree (key_hash);
CREATE INDEX commission_by_account ON commission USING btree (((key_value ->>'from')::text), ((key_value ->> 'to')::text));
CREATE INDEX commission_by_account_type ON commission USING btree (((key_value ->>'from_type')::integer), ((key_value ->> 'to_type')::integer));
CREATE INDEX commission_by_asset ON commission USING btree (((key_value ->>'asset_type')::text), ((key_value ->> 'asset_code')::text), ((key_value ->> 'asset_issuer')::text));

-- +migrate Down

DROP TABLE commission;
