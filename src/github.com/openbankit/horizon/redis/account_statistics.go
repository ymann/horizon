package redis

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"strconv"
	"time"
)

type AccountStatistics struct {
	Account            string
	AssetCode          string
	Balance            int64
	AccountsStatistics map[xdr.AccountType]history.AccountStatistics
}

// Creates new instance of Account Statistics.
func ReadAccountStatistics(account, assetCode string, data map[string]int64, counterparties []xdr.AccountType) *AccountStatistics {
	stats := NewAccountStatistics(account, assetCode, data["balance"], make(map[xdr.AccountType]history.AccountStatistics))

	for _, counterparty := range counterparties {
		accountStat := stats.readFromMap(data, counterparty)
		if accountStat == nil {
			continue
		}
		stats.AccountsStatistics[counterparty] = *accountStat
	}
	return stats
}

func NewAccountStatistics(account, assetCode string, balance int64, accountStats map[xdr.AccountType]history.AccountStatistics) *AccountStatistics {
	return &AccountStatistics{
		Account:            account,
		AssetCode:          assetCode,
		Balance:            balance,
		AccountsStatistics: accountStats,
	}
}

func (stats *AccountStatistics) GetKey() string {
	return GetAccountStatisticsKey(stats.Account, stats.AssetCode)
}

func GetAccountStatisticsKey(account, assetCode string) string {
	return getKey(namespace_account_stats, account, assetCode)
}

func (stats *AccountStatistics) ToArray() []interface{} {
	var result []interface{}
	for _, value := range stats.AccountsStatistics {
		statArray := stats.toArray(&value)
		if len(result) == 0 {
			result = make([]interface{}, 3, len(stats.AccountsStatistics)*len(statArray)+3)
			result[0] = stats.GetKey()
			result[1] = "balance"
			result[2] = stats.Balance
		}
		result = append(result, statArray...)
	}
	return result
}

func (stats *AccountStatistics) toArray(accountStat *history.AccountStatistics) []interface{} {
	counterparty := int(accountStat.CounterpartyType)
	prefix := strconv.Itoa(counterparty)
	return []interface{}{
		prefix + "tp", counterparty,
		prefix + "di", accountStat.DailyIncome,
		prefix + "do", accountStat.DailyOutcome,
		prefix + "wi", accountStat.WeeklyIncome,
		prefix + "wo", accountStat.WeeklyOutcome,
		prefix + "mi", accountStat.MonthlyIncome,
		prefix + "mo", accountStat.MonthlyOutcome,
		prefix + "ai", accountStat.AnnualIncome,
		prefix + "ao", accountStat.AnnualOutcome,
		prefix + "uat", accountStat.UpdatedAt.Unix(),
	}
}

func (stats *AccountStatistics) readFromMap(data map[string]int64, counterparty xdr.AccountType) *history.AccountStatistics {
	prefix := strconv.Itoa(int(counterparty))
	if _, ok := data[prefix+"tp"]; !ok {
		return nil
	}
	return &history.AccountStatistics{
		Account:          stats.Account,
		AssetCode:        stats.AssetCode,
		CounterpartyType: int16(counterparty),
		DailyIncome:      data[prefix+"di"],
		DailyOutcome:     data[prefix+"do"],
		WeeklyIncome:     data[prefix+"wi"],
		WeeklyOutcome:    data[prefix+"wo"],
		MonthlyIncome:    data[prefix+"mi"],
		MonthlyOutcome:   data[prefix+"mo"],
		AnnualIncome:     data[prefix+"ai"],
		AnnualOutcome:    data[prefix+"ao"],
		UpdatedAt:        time.Unix(data[prefix+"uat"], 0),
	}
}
