package helpers

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
)

type AccountStatsGetter func(stats *history.AccountStatistics) int64

func SumAccountStats(stats map[xdr.AccountType]history.AccountStatistics, statsGetter AccountStatsGetter, accountTypes ...xdr.AccountType) int64 {
	sum := int64(0)
	for _, accType := range accountTypes {
		if acc, ok := stats[xdr.AccountType(accType)]; ok {
			sum += statsGetter(&acc)
		}
	}
	return sum
}
