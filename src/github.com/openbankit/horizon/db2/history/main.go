// Package history contains database record definitions useable for
// reading rows from a the history portion of horizon's database
package history

import (
	"time"

	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2"
	sq "github.com/lann/squirrel"
)

// QInterface is a helper struct on which to hang common queries against a history
// portion of the horizon database.
type QInterface interface {
	// Account limits
	// GetAccountLimits returns limits row by account and asset.
	GetAccountLimits(dest interface{}, address string, assetCode string) error
	// Inserts new account limits instance
	CreateAccountLimits(limits AccountLimits) error
	// Updates account's limits
	UpdateAccountLimits(limits AccountLimits) error

	// Account statistics
	// GetStatisticsByAccountAndAsset selects rows from `account_statistics` by address and asset code
	// Now is used to clear obsolete stats
	GetStatisticsByAccountAndAsset(dest map[xdr.AccountType]AccountStatistics, addy string, assetCode string, now time.Time) error
	// Returns account's statistics for assetCode and counterparty type
	GetAccountStatistics(address string, assetCode string, counterPartyType xdr.AccountType) (AccountStatistics, error)

	// Asset
	// Returns asset for specified xdr.Asset
	Asset(dest interface{}, asset xdr.Asset) error
	// Returns asset for asset type, code and issuer
	AssetByParams(dest interface{}, assetType int, code string, issuer string) error
	// Deletes asset from db by id
	DeleteAsset(id int64) (bool, error)
	// updates asset
	UpdateAsset(asset *Asset) (bool, error)
	// inserts asset
	InsertAsset(asset *Asset) (err error)

	// Account
	// AccountByAddress loads a row from `history_accounts`, by address
	AccountByAddress(dest interface{}, addy string) error
	// loads a id from `history_accounts`, by address
	AccountIDByAddress(addy string) (int64, error)
	// Update account
	AccountUpdate(account *Account) error

	// Commission
	// selects commission by hash
	CommissionByHash(hash string) (*Commission, error)
	// Inserts new commission
	InsertCommission(commission *Commission) (err error)
	// Deletes commission
	DeleteCommission(hash string) (bool, error)
	// update commission
	UpdateCommission(commission *Commission) (bool, error)
	// get highest weight commission
	GetHighestWeightCommission(keys map[string]CommissionKey) (resultingCommissions []Commission, err error)


	// Tries to get operation by id. If does not exists returns sql.ErrNoRows
	OperationByID(dest interface{}, id int64) error

	// Options
	// Tries to select options by name. If not found, returns nil,nil
	OptionsByName(name string) (*Options, error)
	OptionsInsert(options *Options) (err error)
	OptionsUpdate(options *Options) (bool, error)
	OptionsDelete(name string) (bool, error)
}

// Q is default implementation of QInterface
type Q struct {
	*db2.Repo
}

// TotalOrderID represents the ID portion of rows that are identified by the
// "TotalOrderID".  See total_order_id.go in the `db` package for details.
type TotalOrderID struct {
	ID int64 `db:"id"`
}

// LatestLedger loads the latest known ledger
func (q *Q) LatestLedger(dest interface{}) error {
	return q.GetRaw(dest, `SELECT COALESCE(MAX(sequence), 0) FROM history_ledgers`)
}

// OldestOutdatedLedgers populates a slice of ints with the first million
// outdated ledgers, based upon the provided `currentVersion` number
func (q *Q) OldestOutdatedLedgers(dest interface{}, currentVersion int) error {
	return q.SelectRaw(dest, `
		SELECT sequence
		FROM history_ledgers
		WHERE importer_version < $1
		ORDER BY sequence ASC
		LIMIT 1000000`, currentVersion)
}

// AccountLimits contains limits for account set by the admin of a bank and
// is a row of data from the `account_limits` table
type AccountLimits struct {
	Account         string `db:"address"`
	AssetCode       string `db:"asset_code"`
	MaxOperationOut int64  `db:"max_operation_out"`
	DailyMaxOut     int64  `db:"daily_max_out"`
	MonthlyMaxOut   int64  `db:"monthly_max_out"`
	MaxOperationIn  int64  `db:"max_operation_in"`
	DailyMaxIn      int64  `db:"daily_max_in"`
	MonthlyMaxIn    int64  `db:"monthly_max_in"`
}

// AccountLimitsQ is a helper struct to aid in configuring queries that loads
// slices of AccountLimits structs.
type AccountLimitsQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

type Commission struct {
	TotalOrderID
	KeyHash    string `db:"key_hash"`
	KeyValue   string `db:"key_value"`
	FlatFee    int64  `db:"flat_fee"`
	PercentFee int64  `db:"percent_fee"`
	weight     int
}

type AuditLog struct {
	Id        int64     `db:"id"`
	Actor     string    `db:"actor"`      //public key of the actor, performing task
	Subject   string    `db:"subject"`    //subject to change
	Action    string    `db:"action"`     //action performed on subject
	Meta      string    `db:"meta"`       //meta information about audit event
	CreatedAt time.Time `db:"created_at"` // time log was created
}

type Asset struct {
	Id          int64  `db:"id"`
	Type        int    `db:"type"`
	Code        string `db:"code"`
	Issuer      string `db:"issuer"`
	IsAnonymous bool   `db:"is_anonymous"`
}

// AssetQ is a helper struct to aid in configuring queries that loads
// slices of Assets.
type AssetQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}
