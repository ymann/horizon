package history

import (
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/toid"
	"github.com/guregu/null"
	sq "github.com/lann/squirrel"
	"time"
)

// Transaction is a row of data from the `history_transactions` table
type Transaction struct {
	TotalOrderID
	TransactionHash  string      `db:"transaction_hash"`
	LedgerSequence   int32       `db:"ledger_sequence"`
	LedgerCloseTime  time.Time   `db:"ledger_close_time"`
	ApplicationOrder int32       `db:"application_order"`
	Account          string      `db:"account"`
	AccountSequence  string      `db:"account_sequence"`
	FeePaid          int32       `db:"fee_paid"`
	OperationCount   int32       `db:"operation_count"`
	TxEnvelope       string      `db:"tx_envelope"`
	TxResult         string      `db:"tx_result"`
	TxMeta           string      `db:"tx_meta"`
	TxFeeMeta        string      `db:"tx_fee_meta"`
	SignatureString  string      `db:"signatures"`
	MemoType         string      `db:"memo_type"`
	Memo             null.String `db:"memo"`
	ValidAfter       null.Int    `db:"valid_after"`
	ValidBefore      null.Int    `db:"valid_before"`
	CreatedAt        time.Time   `db:"created_at"`
	UpdatedAt        time.Time   `db:"updated_at"`

	rawAccountSequence int64
	rawSignatures interface{}
	rawTimeBounds interface{}
}

func NewTransaction(id int64, hash string, ledgerSeq, index int32, account string, accountSeq int64, fee, opCount int32,
	envelope, result, meta, feeMeta string, rawSignatures, rawTimeBounds interface{}, memoType string,
	memo null.String, created, updated time.Time) *Transaction {
	return &Transaction{
		TotalOrderID: TotalOrderID{
			ID: id,
		},
		TransactionHash:  hash,
		LedgerSequence:   ledgerSeq,
		ApplicationOrder: index,
		Account:          account,
		rawAccountSequence:  accountSeq,
		FeePaid:          fee,
		OperationCount:   opCount,
		TxEnvelope:       envelope,
		TxResult:         result,
		TxMeta:           meta,
		rawSignatures:    rawSignatures,
		rawTimeBounds:    rawTimeBounds,
		MemoType:         memoType,
		Memo:             memo,
		CreatedAt:        created,
		UpdatedAt:        updated,
	}
}

// Returns array of params to be inserted/updated
func (tx *Transaction) GetParams() []interface{} {
	return []interface{}{
		tx.ID,
		tx.TransactionHash,
		tx.LedgerSequence,
		tx.ApplicationOrder,
		tx.Account,
		tx.rawAccountSequence,
		tx.FeePaid,
		tx.OperationCount,
		tx.TxEnvelope,
		tx.TxResult,
		tx.TxMeta,
		tx.TxFeeMeta,
		tx.rawSignatures,
		tx.rawTimeBounds,
		tx.MemoType,
		tx.Memo,
		tx.CreatedAt,
		tx.UpdatedAt,
	}
}

// Returns hash of the object. Must be immutable
func (tx *Transaction) Hash() uint64 {
	return uint64(tx.ID)
}

// Returns true if this and other are equals
func (tx *Transaction) Equals(rawOther interface{}) bool {
	other, ok := rawOther.(*Transaction)
	if !ok {
		return false
	}
	return tx.ID == other.ID
}

// TransactionsQ is a helper struct to aid in configuring queries that loads
// slices of transaction structs.
type TransactionsQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

// TransactionByHash is a query that loads a single row from the
// `history_transactions` table based upon the provided hash.
func (q *Q) TransactionByHash(dest interface{}, hash string) error {
	sql := selectTransaction.
		Limit(1).
		Where("ht.transaction_hash = ?", hash)

	return q.Get(dest, sql)
}

// Transactions provides a helper to filter rows from the `history_transactions`
// table with pre-defined filters.  See `TransactionsQ` methods for the
// available filters.
func (q *Q) Transactions() *TransactionsQ {
	return &TransactionsQ{
		parent: q,
		sql:    selectTransaction,
	}
}

// ForAccount filters the transactions collection to a specific account
func (q *TransactionsQ) ForAccount(aid string) *TransactionsQ {
	var account Account
	q.Err = q.parent.AccountByAddress(&account, aid)
	if q.Err != nil {
		return q
	}

	q.sql = q.sql.
		Join("history_transaction_participants htp ON htp.history_transaction_id = ht.id").
		Where("htp.history_account_id = ?", account.ID)

	return q
}

// ForLedger filters the query to a only transactions in a specific ledger,
// specified by its sequence.
func (q *TransactionsQ) ForLedger(seq int32) *TransactionsQ {
	var ledger Ledger
	q.Err = q.parent.LedgerBySequence(&ledger, seq)
	if q.Err != nil {
		return q
	}

	start := toid.ID{LedgerSequence: seq}
	end := toid.ID{LedgerSequence: seq + 1}
	q.sql = q.sql.Where(
		"ht.id >= ? AND ht.id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *TransactionsQ) Page(page db2.PageQuery) *TransactionsQ {
	if q.Err != nil {
		return q
	}

	q.sql, q.Err = page.ApplyTo(q.sql, "ht.id")
	return q
}

func (q *TransactionsQ) ClosedAt(closeAt db2.CloseAtQuery) *TransactionsQ {
	if q.Err != nil {
		return q
	}
	q.sql, q.Err = closeAt.ApplyTo(q.sql, "hl.closed_at")
	return q
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *TransactionsQ) Select(dest interface{}) error {
	if q.Err != nil {
		return q.Err
	}

	q.Err = q.parent.Select(dest, q.sql)
	return q.Err
}

var selectTransaction = sq.Select(
	"ht.id, " +
		"ht.transaction_hash, " +
		"ht.ledger_sequence, " +
		"ht.application_order, " +
		"ht.account, " +
		"ht.account_sequence, " +
		"ht.fee_paid, " +
		"ht.operation_count, " +
		"ht.tx_envelope, " +
		"ht.tx_result, " +
		"ht.tx_meta, " +
		"ht.tx_fee_meta, " +
		"ht.created_at, " +
		"ht.updated_at, " +
		"array_to_string(ht.signatures, ',') AS signatures, " +
		"ht.memo_type, " +
		"ht.memo, " +
		"lower(ht.time_bounds) AS valid_after, " +
		"upper(ht.time_bounds) AS valid_before, " +
		"hl.closed_at AS ledger_close_time").
	From("history_transactions ht").
	LeftJoin("history_ledgers hl ON ht.ledger_sequence = hl.sequence")

var TransactionInsert = sq.Insert("history_transactions").Columns(
	"id",
	"transaction_hash",
	"ledger_sequence",
	"application_order",
	"account",
	"account_sequence",
	"fee_paid",
	"operation_count",
	"tx_envelope",
	"tx_result",
	"tx_meta",
	"tx_fee_meta",
	"signatures",
	"time_bounds",
	"memo_type",
	"memo",
	"created_at",
	"updated_at",
)

var TransactionParticipantInsert = sq.Insert("history_transaction_participants").Columns(
	"history_transaction_id",
	"history_account_id",
)
