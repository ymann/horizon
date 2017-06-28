package history

import (
	"github.com/openbankit/horizon/db2"
	"github.com/guregu/null"
	sq "github.com/lann/squirrel"
	"time"
)

// Ledger is a row of data from the `history_ledgers` table
type Ledger struct {
	TotalOrderID
	Sequence           uint32      `db:"sequence"`
	ImporterVersion    int32       `db:"importer_version"`
	LedgerHash         string      `db:"ledger_hash"`
	PreviousLedgerHash null.String `db:"previous_ledger_hash"`
	TransactionCount   uint32      `db:"transaction_count"`
	OperationCount     uint32      `db:"operation_count"`
	ClosedAt           time.Time   `db:"closed_at"`
	CreatedAt          time.Time   `db:"created_at"`
	UpdatedAt          time.Time   `db:"updated_at"`
	TotalCoins         int64       `db:"total_coins"`
	FeePool            int64       `db:"fee_pool"`
	BaseFee            uint32      `db:"base_fee"`
	BaseReserve        uint32      `db:"base_reserve"`
	MaxTxSetSize       uint32      `db:"max_tx_set_size"`
}

func NewLedger(importerV int32, id int64, sequence uint32, hash string, previousLedgerHash null.String, totalCoins, feePool int64,
	baseFee, baseReserve, maxTxSetSize uint32, close, create, update time.Time, txs, ops uint32) *Ledger {
	return &Ledger{
		ImporterVersion: importerV,
		TotalOrderID: TotalOrderID{
			ID: id,
		},
		Sequence:           sequence,
		LedgerHash:         hash,
		PreviousLedgerHash: previousLedgerHash,
		TotalCoins:         totalCoins,
		FeePool:            feePool,
		BaseFee:            baseFee,
		BaseReserve:        baseReserve,
		MaxTxSetSize:       maxTxSetSize,
		ClosedAt:           close,
		CreatedAt:          create,
		UpdatedAt:          update,
		TransactionCount:   txs,
		OperationCount:     ops,
	}
}

// Returns array of params to be inserted/updated
func (ledger *Ledger) GetParams() []interface{} {
	return []interface{}{
		ledger.ImporterVersion,
		ledger.ID,
		ledger.Sequence,
		ledger.LedgerHash,
		ledger.PreviousLedgerHash,
		ledger.TotalCoins,
		ledger.FeePool,
		ledger.BaseFee,
		ledger.BaseReserve,
		ledger.MaxTxSetSize,
		ledger.ClosedAt,
		ledger.CreatedAt,
		ledger.UpdatedAt,
		ledger.TransactionCount,
		ledger.OperationCount,
	}
}

// Returns hash of the object. Must be immutable
func (ledger *Ledger) Hash() uint64 {
	return uint64(ledger.ID)
}

// Returns true if this and other are equals
func (ledger *Ledger) Equals(rawOther interface{}) bool {
	other, ok := rawOther.(*Ledger)
	if !ok {
		return false
	}
	return ledger.ID == other.ID
}

// LedgersQ is a helper struct to aid in configuring queries that loads
// slices of Ledger structs.
type LedgersQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

// LedgerBySequence loads the single ledger at `seq` into `dest`
func (q *Q) LedgerBySequence(dest interface{}, seq int32) error {
	sql := selectLedger.
		Limit(1).
		Where("sequence = ?", seq)

	return q.Get(dest, sql)
}

// Ledgers provides a helper to filter rows from the `history_ledgers` table
// with pre-defined filters.  See `LedgersQ` methods for the available filters.
func (q *Q) Ledgers() *LedgersQ {
	return &LedgersQ{
		parent: q,
		sql:    selectLedger,
	}
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *LedgersQ) Page(page db2.PageQuery) *LedgersQ {
	if q.Err != nil {
		return q
	}

	q.sql, q.Err = page.ApplyTo(q.sql, "hl.id")
	return q
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *LedgersQ) Select(dest interface{}) error {
	if q.Err != nil {
		return q.Err
	}

	q.Err = q.parent.Select(dest, q.sql)
	return q.Err
}

var selectLedger = sq.Select(
	"hl.id",
	"hl.sequence",
	"hl.importer_version",
	"hl.ledger_hash",
	"hl.previous_ledger_hash",
	"hl.transaction_count",
	"hl.operation_count",
	"hl.closed_at",
	"hl.created_at",
	"hl.updated_at",
	"hl.total_coins",
	"hl.fee_pool",
	"hl.base_fee",
	"hl.base_reserve",
	"hl.max_tx_set_size",
).From("history_ledgers hl")

var LedgerInsert = sq.Insert("history_ledgers").Columns(
	"importer_version",
	"id",
	"sequence",
	"ledger_hash",
	"previous_ledger_hash",
	"total_coins",
	"fee_pool",
	"base_fee",
	"base_reserve",
	"max_tx_set_size",
	"closed_at",
	"created_at",
	"updated_at",
	"transaction_count",
	"operation_count",
)
