package history

import (
	"time"

	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/helpers"
	"github.com/openbankit/horizon/log"
	sq "github.com/lann/squirrel"
)

// AccountStatistics is a row of data from the `account_statistics` table
type AccountStatistics struct {
	Account          string    `db:"address"`
	AssetCode        string    `db:"asset_code"`
	CounterpartyType int16     `db:"counterparty_type"`
	DailyIncome      int64     `db:"daily_income"`
	DailyOutcome     int64     `db:"daily_outcome"`
	WeeklyIncome     int64     `db:"weekly_income"`
	WeeklyOutcome    int64     `db:"weekly_outcome"`
	MonthlyIncome    int64     `db:"monthly_income"`
	MonthlyOutcome   int64     `db:"monthly_outcome"`
	AnnualIncome     int64     `db:"annual_income"`
	AnnualOutcome    int64     `db:"annual_outcome"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// Returns array of params to be inserted/updated
func (stats *AccountStatistics) GetParams() []interface{} {
	return []interface{}{
		stats.Account,
		stats.AssetCode,
		int16(stats.CounterpartyType),
		stats.DailyIncome,
		stats.DailyOutcome,
		stats.WeeklyIncome,
		stats.WeeklyOutcome,
		stats.MonthlyIncome,
		stats.MonthlyOutcome,
		stats.AnnualIncome,
		stats.AnnualOutcome,
		stats.UpdatedAt,
	}
}

// Returns hash of the object. Must be immutable
func (stats *AccountStatistics) Hash() uint64 {
	initialOddNumber := uint64(19)
	result := initialOddNumber + helpers.StringHashCode(stats.Account)
	result = result*uint64(29) + helpers.StringHashCode(stats.AssetCode)
	return result*uint64(31) + uint64(stats.CounterpartyType)
}

// Returns true if this and other are equals
func (stats *AccountStatistics) Equals(rawOther interface{}) bool {
	other, ok := rawOther.(*AccountStatistics)
	if !ok {
		return false
	}
	return stats.Account == other.Account && stats.AssetCode == other.AssetCode && stats.CounterpartyType == other.CounterpartyType
}

func (stats *AccountStatistics) GetKeyParams() []interface{} {
	return []interface{}{
		stats.Account,
		stats.AssetCode,
		stats.CounterpartyType,
	}
}

// AccountStatisticsQ is a helper struct to aid in configuring queries that loads
// slices of Ledger structs.
type AccountStatisticsQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

func NewAccountStatistics(account, assetCode string, counterparty xdr.AccountType) AccountStatistics {
	return AccountStatistics{
		Account:          account,
		AssetCode:        assetCode,
		CounterpartyType: int16(counterparty),
	}
}

func (stats *AccountStatistics) Update(delta int64, receivedAt time.Time, now time.Time, isIncome bool) {
	log.WithFields(log.F{
		"service":    "account_statistics",
		"delta":      delta,
		"receivedAt": receivedAt,
		"now":        now,
		"isIncome":   isIncome,
	}).Debug("Updating")
	if isIncome {
		stats.AddIncome(delta, receivedAt, now)
	} else {
		stats.AddOutcome(delta, receivedAt, now)
	}
}

func (stats *AccountStatistics) AddIncome(income int64, receivedAt time.Time, now time.Time) {
	if receivedAt.Year() != now.Year() {
		return
	}
	stats.AnnualIncome = stats.AnnualIncome + income

	if receivedAt.Month() != now.Month() {
		return
	}
	stats.MonthlyIncome = stats.MonthlyIncome + income

	if !helpers.SameWeek(receivedAt, now) {
		return
	}
	stats.WeeklyIncome = stats.WeeklyIncome + income

	if receivedAt.Day() != now.Day() {
		return
	}
	stats.DailyIncome = stats.DailyIncome + income
}

func (stats *AccountStatistics) AddOutcome(outcome int64, performedAt time.Time, now time.Time) {
	if performedAt.Year() != now.Year() {
		return
	}
	stats.AnnualOutcome = stats.AnnualOutcome + outcome

	if performedAt.Month() != now.Month() {
		return
	}
	stats.MonthlyOutcome = stats.MonthlyOutcome + outcome

	if !helpers.SameWeek(performedAt, now) {
		return
	}
	stats.WeeklyOutcome = stats.WeeklyOutcome + outcome

	if performedAt.Day() != now.Day() {
		return
	}
	stats.DailyOutcome = stats.DailyOutcome + outcome
}

// GetAccountStatistics returns account_statistics row by account, asset and counterparty type.
func (q *Q) GetAccountStatistics(address string, assetCode string, counterPartyType xdr.AccountType) (AccountStatistics, error) {
	sql := selectAccountStatisticsTemplate.Limit(1).Where("a.address = ? AND a.asset_code = ? AND a.counterparty_type = ?",
		address,
		assetCode,
		int16(counterPartyType),
	)

	var stats AccountStatistics
	err := q.Get(&stats, sql)
	return stats, err
}

// GetStatisticsByAccount selects rows from `account_statistics` by address
func (q *Q) GetStatisticsByAccount(dest *[]AccountStatistics, addy string) error {
	sql := selectAccountStatisticsTemplate.Where("a.address = ?", addy)
	var stats []AccountStatistics
	err := q.Select(&stats, sql)

	if err == nil {
		now := time.Now()
		for _, stat := range stats {
			// Erase obsolete data from result. Don't save, to avoid conflicts with ingester's thread
			stat.ClearObsoleteStats(now)
		}
		*dest = stats
	}

	return err
}

// GetStatisticsByAccountAndAsset selects rows from `account_statistics` by address and asset code
func (q *Q) GetStatisticsByAccountAndAsset(dest map[xdr.AccountType]AccountStatistics, addy string, assetCode string, now time.Time) error {
	sql := selectAccountStatisticsTemplate.Where("a.address = ? AND a.asset_code = ?", addy, assetCode)
	var stats []AccountStatistics
	err := q.Select(&stats, sql)
	if err != nil {
		return err
	}

	for _, stat := range stats {
		// Erase obsolete data from result. Don't save, to avoid conflicts with ingester's thread
		stat.ClearObsoleteStats(now)
		dest[xdr.AccountType(stat.CounterpartyType)] = stat
	}

	return nil
}

// ClearObsoleteStats checks last update time and erases obsolete data
func (stats *AccountStatistics) ClearObsoleteStats(now time.Time) {
	log.WithField("now", now).WithField("updated_at", stats.UpdatedAt).Debug("Clearing obsolete")
	isYear := stats.UpdatedAt.Year() < now.Year()
	if isYear {
		stats.AnnualIncome = 0
		stats.AnnualOutcome = 0
	}
	isMonth := isYear || stats.UpdatedAt.Month() < now.Month()
	if isMonth {

		stats.MonthlyIncome = 0
		stats.MonthlyOutcome = 0
	}
	isWeek := isMonth || !helpers.SameWeek(stats.UpdatedAt, now)
	if isWeek {
		stats.WeeklyIncome = 0
		stats.WeeklyOutcome = 0
	}
	isDay := isWeek || stats.UpdatedAt.Day() < now.Day()
	log.WithFields(
		log.F{
			"service":           "account_statistics",
			"account":           stats.Account,
			"asset":             stats.AssetCode,
			"counterparty_type": stats.CounterpartyType,
			"year":              isYear,
			"month":             isMonth,
			"week":              isWeek,
			"day":               isDay,
			"now":               now.String(),
			"updated":           stats.UpdatedAt.String(),
		}).Debug("Erasing obsolete stats")
	if isDay {
		stats.DailyIncome = 0
		stats.DailyOutcome = 0

		stats.UpdatedAt = now
	}
}

// SelectAccountStatisticsTemplate is a prepared statement for SELECT from the account_statistics
var selectAccountStatisticsTemplate = sq.Select("a.*").From("account_statistics a")

// CreateAccountStatisticsTemplate is a prepared statement for insertion into the account_statistics
var AccountStatisticsInsert = sq.Insert("account_statistics").Columns(
	"address",
	"asset_code",
	"counterparty_type",
	"daily_income",
	"daily_outcome",
	"weekly_income",
	"weekly_outcome",
	"monthly_income",
	"monthly_outcome",
	"annual_income",
	"annual_outcome",
	"updated_at",
)

var AccountStatisticsUpdateParams = []string{
	"address",
	"asset_code",
	"counterparty_type",
	"daily_income",
	"daily_outcome",
	"weekly_income",
	"weekly_outcome",
	"monthly_income",
	"monthly_outcome",
	"annual_income",
	"annual_outcome",
	"updated_at",
}

var AccountStatisticsUpdateWhere = "address = ? AND asset_code = ? AND counterparty_type = ?"

var updateAccountStatisticsTemplate = sq.Update("account_statistics")
