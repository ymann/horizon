package session

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"database/sql"
	"time"
)

func (is *Session) ingestPayment(sourceAddress, destAddress string, sourceAmount, destAmount xdr.Int64, sourceAsset, destAsset string) error {

	sourceAccount, err := is.Ingestion.HistoryAccountCache.Get(sourceAddress)
	if err != nil {
		return err
	}

	destAccount, err := is.Ingestion.HistoryAccountCache.Get(destAddress)
	isDestNew := false
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		isDestNew = true
	}

	if isDestNew {
		destAccount = history.NewAccount(is.Cursor.OperationID(), destAddress, xdr.AccountTypeAccountAnonymousUser)
		err = is.Ingestion.Account(destAccount, true, &destAsset, &sourceAccount.AccountType)
		if err != nil {
			log.Error("Failed to ingest anonymous account created by payment!")
			return err
		}
	}

	ledgerCloseTime := time.Unix(is.Cursor.Ledger().CloseTime, 0).Local()
	now := time.Now()
	err = is.Ingestion.UpdateStatistics(sourceAddress, sourceAsset, destAccount.AccountType, int64(sourceAmount), ledgerCloseTime, now, false)
	if err != nil {
		return err
	}

	return is.Ingestion.UpdateStatistics(destAddress, destAsset, sourceAccount.AccountType, int64(destAmount), ledgerCloseTime, now, true)
}
