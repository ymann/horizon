-- +migrate Up

ALTER TABLE account_limits RENAME COLUMN max_operation TO max_operation_out;
ALTER TABLE account_limits RENAME COLUMN daily_turnover TO daily_max_out;
ALTER TABLE account_limits RENAME COLUMN monthly_turnover TO monthly_max_out;
ALTER TABLE account_limits
  ADD COLUMN max_operation_in bigint NOT NULL DEFAULT -1,
  ADD COLUMN daily_max_in bigint NOT NULL DEFAULT -1,
  ADD COLUMN monthly_max_in bigint NOT NULL DEFAULT -1;

-- +migrate Down

ALTER TABLE account_limits RENAME COLUMN max_operation_out TO max_operation;
ALTER TABLE account_limits RENAME COLUMN daily_max_out TO daily_turnover;
ALTER TABLE account_limits RENAME COLUMN monthly_max_out TO monthly_turnover;
ALTER TABLE account_limits
  DROP COLUMN max_operation_in,
  DROP COLUMN daily_max_in,
  DROP COLUMN monthly_max_in;
