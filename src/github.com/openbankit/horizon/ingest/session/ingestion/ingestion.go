package ingestion

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/db2/sqx"
	"github.com/guregu/null"
	sq "github.com/lann/squirrel"
	"database/sql"
)

// Account ingests the provided account data into a new row in the
// `history_accounts` table
func (ingest *Ingestion) Account(account *history.Account, skipCheck bool, createdByAsset *string, createdByCounterparty *xdr.AccountType) error {

	if skipCheck {
		_, err := ingest.HistoryAccountCache.Get(account.Address)

		if err != sql.ErrNoRows {
			return err
		}
	}

	err := ingest.accounts.Insert(account)
	if err != nil {
		return err
	}

	ingest.HistoryAccountCache.Add(account.Address, account)

	// if created by data exists add statistics
	if createdByAsset != nil && createdByCounterparty != nil {
		ingest.statisticsCache.AddWithParams(account.Address, *createdByAsset, *createdByCounterparty, nil)
	}

	return nil
}

// Effect adds a new row into the `history_effects` table.
func (ingest *Ingestion) Effect(aid int64, opid int64, order int, typ history.EffectType, details interface{}) error {
	djson, err := json.Marshal(details)
	if err != nil {
		return err
	}

	return ingest.effects.Insert(history.NewEffect(aid, opid, int32(order), typ, djson))
}

// Ledger adds a ledger to the current ingestion
func (ingest *Ingestion) Ledger(
	id int64,
	header *core.LedgerHeader,
	txs int,
	ops int,
) error {

	ledger := history.NewLedger(
		int32(ingest.CurrentVersion),
		id,
		header.Sequence,
		header.LedgerHash,
		null.NewString(header.PrevHash, header.Sequence > 1),
		int64(header.Data.TotalCoins),
		int64(header.Data.FeePool),
		uint32(header.Data.BaseFee),
		uint32(header.Data.BaseReserve),
		uint32(header.Data.MaxTxSetSize),
		time.Unix(header.CloseTime, 0).UTC(),
		time.Now().UTC(),
		time.Now().UTC(),
		uint32(txs),
		uint32(ops),
	)

	return ingest.ledgers.Insert(ledger)
}

// Operation ingests the provided operation data into a new row in the
// `history_operations` table
func (ingest *Ingestion) Operation(
	id int64,
	txid int64,
	order int32,
	source xdr.AccountId,
	typ xdr.OperationType,
	details map[string]interface{},

) error {
	djson, err := json.Marshal(details)
	if err != nil {
		return err
	}

	operation := history.NewOperation(id, txid, order, source.Address(), typ, djson)
	err = ingest.operations.Insert(operation)
	if err != nil {
		return err
	}

	return nil
}

// OperationParticipants ingests the provided accounts `aids` as participants of
// operation with id `op`, creating a new row in the
// `history_operation_participants` table.
func (ingest *Ingestion) OperationParticipants(op int64, aids []int64) error {
	for _, aid := range aids {
		err := ingest.operation_participants.Insert(history.NewParticipant(op, aid))
		if err != nil {
			return err
		}
	}

	return nil
}

// Transaction ingests the provided transaction data into a new row in the
// `history_transactions` table
func (ingest *Ingestion) Transaction(
	id int64,
	tx *core.Transaction,
	fee *core.TransactionFee,
) error {

	historyTx := history.NewTransaction(
		id,
		tx.TransactionHash,
		tx.LedgerSequence,
		tx.Index,
		tx.SourceAddress(),
		tx.Sequence(),
		tx.Fee(),
		int32(len(tx.Envelope.Tx.Operations)),
		tx.EnvelopeXDR(),
		tx.ResultXDR(),
		tx.ResultMetaXDR(),
		fee.ChangesXDR(),
		sqx.StringArray(tx.Base64Signatures()),
		formatTimeBounds(tx.Envelope.Tx.TimeBounds),
		tx.MemoType(),
		tx.Memo(),
		time.Now().UTC(),
		time.Now().UTC(),
	)

	return ingest.transactions.Insert(historyTx)
}

// TransactionParticipants ingests the provided account ids as participants of
// transaction with id `tx`, creating a new row in the
// `history_transaction_participants` table.
func (ingest *Ingestion) TransactionParticipants(tx int64, aids []int64) error {
	for _, aid := range aids {
		participant := history.NewParticipant(tx, aid)
		err := ingest.transaction_participants.Insert(participant)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ingest *Ingestion) clearRange(start int64, end int64, table string, idCol string) error {
	del := sq.Delete(table).Where(
		fmt.Sprintf("%s >= ? AND %s < ?", idCol, idCol),
		start,
		end,
	)
	_, err := ingest.DB.Exec(del)
	return err
}

func (ingest *Ingestion) commit() error {
	err := ingest.DB.Commit()
	if err != nil {
		return err
	}

	return nil
}
