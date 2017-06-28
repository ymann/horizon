package history

import (
	"encoding/json"
	"strings"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/toid"
	"github.com/go-errors/errors"
	"github.com/guregu/null"
	sq "github.com/lann/squirrel"
	"time"
)

// Operation is a row of data from the `history_operations` table
type Operation struct {
	TotalOrderID
	TransactionID    int64             `db:"transaction_id"`
	TransactionHash  string            `db:"transaction_hash"`
	ApplicationOrder int32             `db:"application_order"`
	Type             xdr.OperationType `db:"type"`
	DetailsString    null.String       `db:"details"`
	SourceAccount    string            `db:"source_account"`
	ClosedAt         time.Time         `db:"closed_at"`

	rawDetails []byte
}

func NewOperation(id, txID int64, appOrder int32, sourceAccount string, typ xdr.OperationType, details []byte) *Operation {
	return &Operation{
		TotalOrderID: TotalOrderID{
			ID: id,
		},
		TransactionID:    txID,
		ApplicationOrder: appOrder,
		SourceAccount:    sourceAccount,
		Type:             typ,
		rawDetails:       details,
	}
}

// Returns array of params to be inserted/updated
func (operation *Operation) GetParams() []interface{} {
	return []interface{}{
		operation.ID,
		operation.TransactionID,
		operation.ApplicationOrder,
		operation.SourceAccount,
		operation.Type,
		operation.rawDetails,
	}
}

// Returns hash of the object. Must be immutable
func (operation *Operation) Hash() uint64 {
	return uint64(operation.ID)
}

// Returns true if this and other are equals
func (operation *Operation) Equals(rawOther interface{}) bool {
	other, ok := rawOther.(*Operation)
	if !ok {
		return false
	}
	return operation.ID == other.ID
}

type Participant struct {
	ActionID  int64
	AccountID int64
}

func NewParticipant(actionID, account int64) *Participant {
	return &Participant{
		ActionID:  actionID,
		AccountID: account,
	}
}

// Returns array of params to be inserted/updated
func (participant *Participant) GetParams() []interface{} {
	return []interface{}{
		participant.ActionID,
		participant.AccountID,
	}
}

// Returns hash of the object. Must be immutable
func (participant *Participant) Hash() uint64 {
	result := uint64(17) + uint64(participant.AccountID)
	return result*uint64(29) + uint64(participant.ActionID)
}

// Returns true if this and other are equals
func (participant *Participant) Equals(rawOther interface{}) bool {
	other, ok := rawOther.(*Participant)
	if !ok {
		return false
	}
	return participant.ActionID == other.ActionID && participant.AccountID == other.AccountID
}

// OperationsQ is a helper struct to aid in configuring queries that loads
// slices of Operation structs.
type OperationsQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

// UnmarshalDetails unmarshals the details of this operation into `dest`
func (r *Operation) UnmarshalDetails(dest interface{}) error {
	if !r.DetailsString.Valid {
		return nil
	}

	err := json.Unmarshal([]byte(r.DetailsString.String), &dest)
	if err != nil {
		err = errors.Wrap(err, 1)
	}

	return err
}

// Operations provides a helper to filter the operations table with pre-defined
// filters.  See `OperationsQ` for the available filters.
func (q *Q) Operations() *OperationsQ {
	return &OperationsQ{
		parent: q,
		sql:    selectOperation,
	}
}

// OperationByID loads a single operation with `id` into `dest`
func (q *Q) OperationByID(dest interface{}, id int64) error {
	sql := selectOperation.
		Limit(1).
		Where("hop.id = ?", id)

	return q.Get(dest, sql)
}

// ForAccount filters the operations collection to a specific account
func (q *OperationsQ) ForAccount(aid string) *OperationsQ {
	var account Account
	q.Err = q.parent.AccountByAddress(&account, aid)
	if q.Err != nil {
		return q
	}

	q.sql = q.sql.Join(
		"history_operation_participants hopp ON "+
			"hopp.history_operation_id = hop.id",
	).Where("hopp.history_account_id = ?", account.ID)

	return q
}

// ForAccounts filters the operations collection to a specific account
func (q *OperationsQ) ForAccounts(ids []string) *OperationsQ {
	var accounts []Account
	q.Err = q.parent.AccountsByAddresses(&accounts, ids)
	if q.Err != nil {
		return q
	}
	// if len(accounts) < 1 {
	// 	return  errors.New("Empty request")
	// }
	addrInterface := make([]interface{}, len(accounts))
	for i, v := range accounts {
    	addrInterface[i] = v.ID
	}

	q.sql = q.sql.Join(
		"history_operation_participants hopp ON "+
			"hopp.history_operation_id = hop.id",
	).Where("hopp.history_account_id in (?"+ strings.Repeat(",?", len(accounts)-1) +")", addrInterface...)

	return q
}


// ForLedger filters the query to a only operations in a specific ledger,
// specified by its sequence.
func (q *OperationsQ) ForLedger(seq int32) *OperationsQ {
	var ledger Ledger
	q.Err = q.parent.LedgerBySequence(&ledger, seq)
	if q.Err != nil {
		return q
	}

	start := toid.ID{LedgerSequence: seq}
	end := toid.ID{LedgerSequence: seq + 1}
	q.sql = q.sql.Where(
		"hop.id >= ? AND hop.id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// ForTransaction filters the query to a only operations in a specific
// transaction, specified by the transactions's hex-encoded hash.
func (q *OperationsQ) ForTransaction(hash string) *OperationsQ {
	var tx Transaction
	q.Err = q.parent.TransactionByHash(&tx, hash)
	if q.Err != nil {
		return q
	}

	start := toid.Parse(tx.ID)
	end := start
	end.TransactionOrder++
	q.sql = q.sql.Where(
		"hop.id >= ? AND hop.id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// OnlyPayments filters the query being built to only include operations that
// are in the "payment" class of operations:  CreateAccountOps, Payments, and
// PathPayments.
func (q *OperationsQ) OnlyPayments() *OperationsQ {
	q.sql = q.sql.Where(sq.Eq{"hop.type": []xdr.OperationType{
		xdr.OperationTypePayment,
		xdr.OperationTypePathPayment,
		xdr.OperationTypeExternalPayment,
		xdr.OperationTypePaymentReversal,
	}})
	return q
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *OperationsQ) Page(page db2.PageQuery) *OperationsQ {
	if q.Err != nil {
		return q
	}

	q.sql, q.Err = page.ApplyTo(q.sql, "hop.id")
	return q
}

func (q *OperationsQ) ClosedAt(closeAt db2.CloseAtQuery) *OperationsQ {
	if q.Err != nil {
		return q
	}
	q.sql, q.Err = closeAt.ApplyTo(q.sql, "hl.closed_at")
	return q
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *OperationsQ) Select(dest interface{}) error {
	if q.Err != nil {
		return q.Err
	}

	q.Err = q.parent.Select(dest, q.sql)
	return q.Err
}

var selectOperation = sq.Select(
	"hop.id, " +
		"hop.transaction_id, " +
		"hop.application_order, " +
		"hop.type, " +
		"hop.details, " +
		"hop.source_account, " +
		"ht.transaction_hash, " +
		"hl.closed_at").
	From("history_operations hop").
	LeftJoin("history_transactions ht ON ht.id = hop.transaction_id").
	LeftJoin("history_ledgers hl ON hl.sequence = ht.ledger_sequence")

var OperationInsert = sq.Insert("history_operations").Columns(
	"id",
	"transaction_id",
	"application_order",
	"source_account",
	"type",
	"details",
)

var OperationParticipantInsert = sq.Insert("history_operation_participants").Columns(
	"history_operation_id",
	"history_account_id",
)
