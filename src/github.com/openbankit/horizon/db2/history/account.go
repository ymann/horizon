package history

import (
	"strings"	
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2"
	"github.com/guregu/null"
	sq "github.com/lann/squirrel"
	"encoding/json"
	"github.com/go-errors/errors"
)

// Account is a row of data from the `history_accounts` table
type Account struct {
	TotalOrderID
	Address                string          `db:"address"`
	AccountType            xdr.AccountType `db:"account_type"`
	BlockIncomingPayments  bool            `db:"block_incoming_payments"`
	BlockOutcomingPayments bool            `db:"block_outcoming_payments"`
	LimitedAssets          null.String     `db:"limited_assets"`
}

func NewAccount(id int64, address string, accountType xdr.AccountType) *Account {
	return &Account{
		TotalOrderID: TotalOrderID{
			ID: id,
		},
		Address:     address,
		AccountType: accountType,
	}
}

// UnmarshalDetails unmarshals the details of this effect into `dest`
func (r *Account) UnmarshalLimitedAssets() (map[string]bool, error) {
	result := make(map[string]bool)
	if !r.LimitedAssets.Valid {
		return result, nil
	}

	err := json.Unmarshal([]byte(r.LimitedAssets.String), &result)
	if err != nil {
		err = errors.Wrap(err, 1)
	}

	return result, err
}

func (r *Account) SetLimitedAssets(limitedAssets map[string]bool) (error) {
	data, err := json.Marshal(limitedAssets)
	if err != nil {
		return err
	}
	r.LimitedAssets = null.StringFrom(string(data))
	return nil
}

// Returns array of params to be inserted/updated
func (account *Account) GetParams() []interface{} {
	return []interface{}{
		account.ID,
		account.Address,
		account.AccountType,
		account.BlockIncomingPayments,
		account.BlockOutcomingPayments,
		account.LimitedAssets,
	}
}

// Returns hash of the object. Must be immutable
func (account *Account) Hash() uint64 {
	return uint64(account.ID)
}

// Returns true if this and other are equals
func (account *Account) Equals(rawOther interface{}) bool {
	other, ok := rawOther.(*Account)
	if !ok {
		return false
	}
	return account.ID == other.ID
}

// AccountsQ is a helper struct to aid in configuring queries that loads
// slices of account structs.
type AccountsQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

// Accounts provides a helper to filter rows from the `history_accounts` table
// with pre-defined filters.  See `AccountsQ` methods for the available filters.
func (q *Q) Accounts() *AccountsQ {
	return &AccountsQ{
		parent: q,
		sql:    selectAccount,
	}
}

// AccountByAddress loads a row from `history_accounts`, by address
func (q *Q) AccountByAddress(dest interface{}, addy string) error {
	sql := selectAccount.Limit(1).Where("ha.address = ?", addy)
	return q.Get(dest, sql)
}

// loads a id from `history_accounts`, by address
func (q *Q) AccountIDByAddress(addy string) (int64, error) {
	var id int64
	err := q.GetRaw(&id, `SELECT id FROM history_accounts WHERE address = $1 ORDER BY id DESC`, addy)
	return id, err
}

// AccountsByAddresses loads rows from `history_accounts`, by addresses
func (q *Q) AccountsByAddresses(dest interface{}, addresses []string) error {
	// if len(addresses) < 1 {
	// 	q.Err = errors.New("Empty request")
	// 	return q
	// }
	addrInterface := make([]interface{}, len(addresses))
	for i, v := range addresses {
    	addrInterface[i] = v
	}
	sql := selectAccount.Where("ha.address in (?"+ strings.Repeat(",?", len(addresses)-1) +")", addrInterface...)
	return q.Select(dest, sql)
}


func (q *Q) AccountUpdate(account *Account) error {
	sql := AccountUpdate.SetMap(map[string]interface{}{
		"block_incoming_payments":  account.BlockIncomingPayments,
		"block_outcoming_payments": account.BlockOutcomingPayments,
		"limited_assets":           account.LimitedAssets,
	}).Where("history_accounts.id = ?", account.ID)
	_, err := q.Exec(sql)
	if err != nil {
		errors.Wrap(err, 0)
	}
	return err
}

// AccountByID loads a row from `history_accounts`, by id
func (q *Q) AccountByID(dest interface{}, id int64) error {
	sql := selectAccount.Limit(1).Where("ha.id = ?", id)
	return q.Get(dest, sql)
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *AccountsQ) Page(page db2.PageQuery) *AccountsQ {
	if q.Err != nil {
		return q
	}

	q.sql, q.Err = page.ApplyTo(q.sql, "ha.id")
	return q
}

func (q *AccountsQ) Blocked() *AccountsQ {
	if q.Err != nil {
		return q
	}

	q.sql = q.sql.Where("(block_outcoming_payments = true OR block_outcoming_payments = true)")
	return q
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *AccountsQ) Select(dest interface{}) error {
	if q.Err != nil {
		return q.Err
	}

	q.Err = q.parent.Select(dest, q.sql)
	return q.Err
}

var selectAccount = sq.Select("ha.*").From("history_accounts ha")

var AccountInsert = sq.Insert("history_accounts").Columns(
	"id",
	"address",
	"account_type",
	"block_incoming_payments",
	"block_outcoming_payments",
	"limited_assets",
)

var AccountUpdate = sq.Update("history_accounts")
