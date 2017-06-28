package ingestion

import (
	"database/sql"
	"time"

	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
)

// updateAccountStats updates outcome stats for specified account and asset
func (ingest *Ingestion) UpdateStatistics(address string, assetCode string, counterpartyType xdr.AccountType,
	delta int64, //account balance change
	ledgerClosedAt time.Time, now time.Time,
	income bool, // payment direction
) error {
	isNew := false
	stats, err := ingest.statisticsCache.Get(address, assetCode, counterpartyType)
	if err != nil || stats == nil {
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		isNew = true
		rawStats := history.NewAccountStatistics(address, assetCode, counterpartyType)
		stats = &rawStats
		ingest.statisticsCache.Add(stats)
	} else {
		stats.ClearObsoleteStats(now)
	}

	stats.Update(delta, ledgerClosedAt, now, income)
	stats.UpdatedAt = now

	if isNew {
		err = ingest.statistics.Insert(stats)
	} else {
		err = ingest.statistics.Update(stats)
	}
	return err
}
