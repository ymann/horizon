package history

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/toid"
	"github.com/go-errors/errors"
	"github.com/guregu/null"
	sq "github.com/lann/squirrel"
)

const (
	// account effects

	// EffectAccountCreated effects occur when a new account is created
	EffectAccountCreated EffectType = 0 // from create_account

	// EffectAccountRemoved effects occur when one account is merged into another
	EffectAccountRemoved EffectType = 1 // from merge_account

	// EffectAccountCredited effects occur when an account receives some currency
	EffectAccountCredited EffectType = 2 // from create_account, payment, path_payment, merge_account

	// EffectAccountDebited effects occur when an account sends some currency
	EffectAccountDebited EffectType = 3 // from create_account, payment, path_payment, create_account

	// EffectAccountThresholdsUpdated effects occur when an account changes its
	// multisig thresholds.
	EffectAccountThresholdsUpdated EffectType = 4 // from set_options

	// EffectAccountHomeDomainUpdated effects occur when an account changes its
	// home domain.
	EffectAccountHomeDomainUpdated EffectType = 5 // from set_options

	// EffectAccountFlagsUpdated effects occur when an account changes its
	// account flags, either clearing or setting.
	EffectAccountFlagsUpdated EffectType = 6 // from set_options

	// signer effects

	// EffectSignerCreated occurs when an account gains a signer
	EffectSignerCreated EffectType = 10 // from set_options

	// EffectSignerRemoved occurs when an account loses a signer
	EffectSignerRemoved EffectType = 11 // from set_options

	// EffectSignerUpdated occurs when an account changes the weight of one of its
	// signers.
	EffectSignerUpdated EffectType = 12 // from set_options

	// trustline effects

	// EffectTrustlineCreated occurs when an account trusts an anchor
	EffectTrustlineCreated EffectType = 20 // from change_trust

	// EffectTrustlineRemoved occurs when an account removes struct by setting the
	// limit of a trustline to 0
	EffectTrustlineRemoved EffectType = 21 // from change_trust

	// EffectTrustlineUpdated occurs when an account changes a trustline's limit
	EffectTrustlineUpdated EffectType = 22 // from change_trust, allow_trust

	// EffectTrustlineAuthorized occurs when an anchor has AUTH_REQUIRED flag set
	// to true and it authorizes another account's trustline
	EffectTrustlineAuthorized EffectType = 23 // from allow_trust

	// EffectTrustlineDeauthorized occurs when an anchor revokes access to a asset
	// it issues.
	EffectTrustlineDeauthorized EffectType = 24 // from allow_trust

	// trading effects

	// EffectOfferCreated occurs when an account offers to trade an asset
	EffectOfferCreated EffectType = 30 // from manage_offer, creat_passive_offer

	// EffectOfferRemoved occurs when an account removes an offer
	EffectOfferRemoved EffectType = 31 // from manage_offer, creat_passive_offer, path_payment

	// EffectOfferUpdated occurs when an offer is updated by the offering account.
	EffectOfferUpdated EffectType = 32 // from manage_offer, creat_passive_offer, path_payment

	// EffectTrade occurs when a trade is initiated because of a path payment or
	// offer operation.
	EffectTrade EffectType = 33 // from manage_offer, creat_passive_offer, path_payment

	// data effects

	// EffectDataCreated occurs when an account gets a new data field
	EffectDataCreated EffectType = 40 // from manage_data

	// EffectDataRemoved occurs when an account removes a data field
	EffectDataRemoved EffectType = 41 // from manage_data

	// EffectDataUpdated occurs when an account changes a data field's value
	EffectDataUpdated EffectType = 42 // from manage_data

	// EffectAdminOpPerformed occurs when an admin operation was performed
	EffectAdminOpPerformed EffectType = 43
)

// EffectType is the numeric type for an effect, used as the `type` field in the
// `history_effects` table.
type EffectType int

// Effect is a row of data from the `history_effects` table
type Effect struct {
	HistoryAccountID   int64       `db:"history_account_id"`
	Account            string      `db:"address"`
	HistoryOperationID int64       `db:"history_operation_id"`
	Order              int32       `db:"order"`
	Type               EffectType  `db:"type"`
	DetailsString      null.String `db:"details"`

	rawDetails []byte
}

func NewEffect(accountID int64, operationId int64, order int32, typ EffectType, details []byte) *Effect {
	return &Effect{
		HistoryAccountID:   accountID,
		HistoryOperationID: operationId,
		Order:              order,
		Type:               typ,
		rawDetails:         details,
	}
}

// Returns array of params to be inserted/updated
func (effect *Effect) GetParams() []interface{} {
	return []interface{}{
		effect.HistoryAccountID,
		effect.HistoryOperationID,
		effect.Order,
		effect.Type,
		effect.rawDetails,
	}
}

// Returns hash of the object. Must be immutable
func (effect *Effect) Hash() uint64 {
	initialOddNumber := uint64(19)
	result := initialOddNumber + uint64(effect.HistoryOperationID)
	return result*uint64(29) + uint64(effect.Order)
}

// Returns true if this and other are equals
func (effect *Effect) Equals(rawOther interface{}) bool {
	other, ok := rawOther.(*Effect)
	if !ok {
		return false
	}
	return effect.HistoryOperationID == other.HistoryOperationID && effect.Order == other.Order
}

// EffectsQ is a helper struct to aid in configuring queries that loads
// slices of Ledger structs.
type EffectsQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

// UnmarshalDetails unmarshals the details of this effect into `dest`
func (r *Effect) UnmarshalDetails(dest interface{}) error {
	if !r.DetailsString.Valid {
		return nil
	}

	err := json.Unmarshal([]byte(r.DetailsString.String), &dest)
	if err != nil {
		err = errors.Wrap(err, 1)
	}

	return err
}

// ID returns a lexically ordered id for this effect record
func (r *Effect) ID() string {
	return fmt.Sprintf("%019d-%010d", r.HistoryOperationID, r.Order)
}

// PagingToken returns a cursor for this effect
func (r *Effect) PagingToken() string {
	return fmt.Sprintf("%d-%d", r.HistoryOperationID, r.Order)
}

// Effects provides a helper to filter rows from the `history_effects`
// table with pre-defined filters.  See `TransactionsQ` methods for the
// available filters.
func (q *Q) Effects() *EffectsQ {
	return &EffectsQ{
		parent: q,
		sql:    selectEffect,
	}
}

// ForAccount filters the operations collection to a specific account
func (q *EffectsQ) ForAccount(aid string) *EffectsQ {
	var account Account
	q.Err = q.parent.AccountByAddress(&account, aid)
	if q.Err != nil {
		return q
	}

	q.sql = q.sql.Where("heff.history_account_id = ?", account.ID)

	return q
}

// ForLedger filters the query to only effects in a specific ledger,
// specified by its sequence.
func (q *EffectsQ) ForLedger(seq int32) *EffectsQ {
	var ledger Ledger
	q.Err = q.parent.LedgerBySequence(&ledger, seq)
	if q.Err != nil {
		return q
	}

	start := toid.ID{LedgerSequence: seq}
	end := toid.ID{LedgerSequence: seq + 1}
	q.sql = q.sql.Where(
		"heff.history_operation_id >= ? AND heff.history_operation_id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// ForOperation filters the query to only effects in a specific operation,
// specified by its id.
func (q *EffectsQ) ForOperation(id int64) *EffectsQ {
	start := toid.Parse(id)
	end := start
	end.IncOperationOrder()
	q.sql = q.sql.Where(
		"heff.history_operation_id >= ? AND heff.history_operation_id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// ForOrderBook filters the query to only effects whose details indicate that
// the effect is for a specific asset pair.
func (q *EffectsQ) ForOrderBook(selling, buying xdr.Asset) *EffectsQ {
	q.orderBookFilter(selling, "sold_")
	if q.Err != nil {
		return q
	}
	q.orderBookFilter(buying, "bought_")
	if q.Err != nil {
		return q
	}

	return q
}

// ForTransaction filters the query to only effects in a specific
// transaction, specified by the transactions's hex-encoded hash.
func (q *EffectsQ) ForTransaction(hash string) *EffectsQ {
	var tx Transaction
	q.Err = q.parent.TransactionByHash(&tx, hash)
	if q.Err != nil {
		return q
	}

	start := toid.Parse(tx.ID)
	end := start
	end.TransactionOrder++
	q.sql = q.sql.Where(
		"heff.history_operation_id >= ? AND heff.history_operation_id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// OfType filters the query to only effects of the given type.
func (q *EffectsQ) OfType(typ EffectType) *EffectsQ {
	q.sql = q.sql.Where("heff.type = ?", typ)
	return q
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *EffectsQ) Page(page db2.PageQuery) *EffectsQ {
	if q.Err != nil {
		return q
	}

	op, idx, err := page.CursorInt64Pair(db2.DefaultPairSep)
	if err != nil {
		q.Err = err
		return q
	}

	if idx > math.MaxInt32 {
		idx = math.MaxInt32
	}

	switch page.Order {
	case "asc":
		q.sql = q.sql.
			Where(`(
					 heff.history_operation_id > ?
				OR (
							heff.history_operation_id = ?
					AND heff.order > ?
				))`, op, op, idx).
			OrderBy("heff.history_operation_id asc, heff.order asc")
	case "desc":
		q.sql = q.sql.
			Where(`(
					 heff.history_operation_id < ?
				OR (
							heff.history_operation_id = ?
					AND heff.order < ?
				))`, op, op, idx).
			OrderBy("heff.history_operation_id desc, heff.order desc")
	}

	q.sql = q.sql.Limit(page.Limit)
	return q
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *EffectsQ) Select(dest interface{}) error {
	if q.Err != nil {
		return q.Err
	}

	q.Err = q.parent.Select(dest, q.sql)
	return q.Err
}

// OfType filters the query to only effects of the given type.
func (q *EffectsQ) orderBookFilter(a xdr.Asset, prefix string) {
	var typ, code, iss string
	q.Err = a.Extract(&typ, &code, &iss)
	if q.Err != nil {
		return
	}

	if a.Type == xdr.AssetTypeAssetTypeNative {
		clause := fmt.Sprintf(`
				(heff.details->>'%sasset_type' = ?
		AND heff.details ?? '%sasset_code' = false
		AND heff.details ?? '%sasset_issuer' = false)`, prefix, prefix, prefix)
		q.sql = q.sql.Where(clause, typ)
		return
	}

	clause := fmt.Sprintf(`
		(heff.details->>'%sasset_type' = ?
	AND heff.details->>'%sasset_code' = ?
	AND heff.details->>'%sasset_issuer' = ?)`, prefix, prefix, prefix)
	q.sql = q.sql.Where(clause, typ, code, iss)
}

var selectEffect = sq.
	Select("heff.*, hacc.address").
	From("history_effects heff").
	LeftJoin("history_accounts hacc ON hacc.id = heff.history_account_id")

var EffectInsert = sq.Insert("history_effects").Columns(
	"history_account_id",
	"history_operation_id",
	"\"order\"",
	"type",
	"details",
)
