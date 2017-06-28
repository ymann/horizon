package ingestion

import (
	"github.com/openbankit/horizon/cache"
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/db2/sqx"
	sq "github.com/lann/squirrel"
)

// Ingestion receives write requests from a Session
type Ingestion struct {
	// DB is the sql repo to be used for writing any rows into the horizon
	// database.
	DB                       *db2.Repo
	CurrentVersion           int

	ledgers                  *sqx.BatchInsertBuilder
	transactions             *sqx.BatchInsertBuilder
	transaction_participants *sqx.BatchInsertBuilder
	operations               *sqx.BatchInsertBuilder
	operation_participants   *sqx.BatchInsertBuilder
	effects                  *sqx.BatchInsertBuilder
	accounts                 *sqx.BatchInsertBuilder
	statistics               *sqx.BatchUpdateBuilder

	needFlush []sqx.Flushable

	// cache
	statisticsCache          *cache.AccountStatistics
	HistoryAccountCache      *cache.HistoryAccount
}

func New(db *db2.Repo, historyAccountCache *cache.HistoryAccount, currentVersion int) *Ingestion {
	q := &history.Q{
		Repo: db,
	}
	return &Ingestion{
		DB:                  db,
		CurrentVersion:      currentVersion,
		HistoryAccountCache: historyAccountCache,
		statisticsCache:     cache.NewAccountStatistics(q),
	}
}

func (ingest *Ingestion) HistoryQ() history.QInterface {
	return &history.Q{
		Repo: ingest.DB,
	}
}

// Rollback aborts this ingestions transaction
func (ingest *Ingestion) Rollback() (err error) {
	// recreates all inserters to release memory
	ingest.createInsertBuilders()
	err = ingest.DB.Rollback()
	return
}

// Start makes the ingestion reeady, initializing the insert builders and tx
func (ingest *Ingestion) Start() (err error) {
	err = ingest.DB.Begin()
	if err != nil {
		return
	}

	ingest.createInsertBuilders()

	return
}

// Clear removes data from the ledger
func (ingest *Ingestion) Clear(start int64, end int64) error {

	if start <= 1 {
		del := sq.Delete("history_accounts").Where("id = 1")
		ingest.DB.Exec(del)
	}

	err := ingest.clearRange(start, end, "history_effects", "history_operation_id")
	if err != nil {
		return err
	}
	err = ingest.clearRange(start, end, "history_operation_participants", "history_operation_id")
	if err != nil {
		return err
	}
	err = ingest.clearRange(start, end, "history_operations", "id")
	if err != nil {
		return err
	}
	err = ingest.clearRange(start, end, "history_transaction_participants", "history_transaction_id")
	if err != nil {
		return err
	}
	err = ingest.clearRange(start, end, "history_transactions", "id")
	if err != nil {
		return err
	}
	err = ingest.clearRange(start, end, "history_accounts", "id")
	if err != nil {
		return err
	}
	err = ingest.clearRange(start, end, "history_ledgers", "id")
	if err != nil {
		return err
	}

	return nil
}

// Close finishes the current transaction and finishes this ingestion.
func (ingest *Ingestion) Close() error {
	err := ingest.flushInserters()
	if err != nil {
		return err
	}
	return ingest.commit()
}

// Flush writes the currently buffered rows to the db, and if successful
// starts a new transaction.
func (ingest *Ingestion) Flush() error {
	err := ingest.flushInserters()
	if err != nil {
		return err
	}
	err = ingest.commit()
	if err != nil {
		return err
	}

	return ingest.Start()
}

func (ingest *Ingestion) flushInserters() error {
	for _, flusher := range ingest.needFlush {
		err := flusher.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ingest *Ingestion) createInsertBuilders() {
	ingest.statistics = sqx.BatchUpdate(sqx.BatchInsertFromInsert(ingest.DB, history.AccountStatisticsInsert),
		history.AccountStatisticsUpdateParams, history.AccountStatisticsUpdateWhere)

	ingest.ledgers = sqx.BatchInsertFromInsert(ingest.DB, history.LedgerInsert)

	ingest.accounts = sqx.BatchInsertFromInsert(ingest.DB, history.AccountInsert)

	ingest.transactions = sqx.BatchInsertFromInsert(ingest.DB, history.TransactionInsert)

	ingest.transaction_participants = sqx.BatchInsertFromInsert(ingest.DB, history.TransactionParticipantInsert)

	ingest.operations = sqx.BatchInsertFromInsert(ingest.DB, history.OperationInsert)

	ingest.operation_participants = sqx.BatchInsertFromInsert(ingest.DB, history.OperationParticipantInsert)

	ingest.effects = sqx.BatchInsertFromInsert(ingest.DB, history.EffectInsert)

	ingest.needFlush = []sqx.Flushable{
		ingest.statistics,
		ingest.ledgers,
		ingest.accounts,
		ingest.transactions,
		ingest.transaction_participants,
		ingest.operations,
		ingest.operation_participants,
		ingest.effects,

	}
}
