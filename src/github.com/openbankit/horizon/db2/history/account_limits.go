package history

import sq "github.com/lann/squirrel"

// GetAccountLimits returns limits row by account and asset.
func (q *Q) GetAccountLimits(
	dest interface{},
	address string,
	assetCode string,
) error {
	sql := SelectAccountLimitsTemplate.Where(
		"a.address = ? AND a.asset_code = ?",
		address,
		assetCode,
	)

	err := q.Get(dest, sql)
	return err
}

// GetLimitsByAccount selects rows from `account_statistics` by address
func (q *Q) GetLimitsByAccount(dest *[]AccountLimits, address string) error {
	sql := SelectAccountLimitsTemplate.Where("a.address = ?", address)
	var limits []AccountLimits
	err := q.Select(&limits, sql)

	if err == nil {
		*dest = limits
	}

	return err
}

// CreateAccountLimits inserts new account_limits row
func (q *Q) CreateAccountLimits(limits AccountLimits) error {
	sql := CreateAccountLimitsTemplate.Values(limits.Account, limits.AssetCode,
		limits.MaxOperationOut, limits.DailyMaxOut, limits.MonthlyMaxOut,
		limits.MaxOperationIn, limits.DailyMaxIn, limits.MonthlyMaxIn)
	_, err := q.Exec(sql)

	return err
}

// UpdateAccountLimits updates account_limits row
func (q *Q) UpdateAccountLimits(limits AccountLimits) error {
	sql := UpdateAccountLimitsTemplate.Set("max_operation_out", limits.MaxOperationOut)
	sql = sql.Set("daily_max_out", limits.DailyMaxOut)
	sql = sql.Set("monthly_max_out", limits.MonthlyMaxOut)
	sql = sql.Set("max_operation_in", limits.MaxOperationIn)
	sql = sql.Set("daily_max_in", limits.DailyMaxIn)
	sql = sql.Set("monthly_max_in", limits.MonthlyMaxIn)
	sql = sql.Where("address = ? and asset_code = ?", limits.Account, limits.AssetCode)

	_, err := q.Exec(sql)

	return err
}

// SelectAccountLimitsTemplate is a prepared statement for SELECT from the account_limits
var SelectAccountLimitsTemplate = sq.Select("a.*").From("account_limits a")

// CreateAccountLimitsTemplate is a prepared statement for insertion into the account_limits
var CreateAccountLimitsTemplate = sq.Insert("account_limits").Columns(
	"address",
	"asset_code",
	"max_operation_out",
	"daily_max_out",
	"monthly_max_out",
	"max_operation_in",
	"daily_max_in",
	"monthly_max_in",
)

// UpdateAccountLimitsTemplate is a prepared statement for insertion into the account_limits
var UpdateAccountLimitsTemplate = sq.Update("account_limits")
