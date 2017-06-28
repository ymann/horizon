// Package resource contains the type definitions for all of horizons
// response resources.
package resource

import (
	"time"

	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/db2/history/details"
	"github.com/openbankit/horizon/render/hal"
	"github.com/openbankit/horizon/resource/base"
	"github.com/openbankit/horizon/resource/effects"
	"github.com/openbankit/horizon/resource/operations"
	"golang.org/x/net/context"
)

// AccountTypeNames maps from account type to the string used to represent that type
// in horizon's JSON responses
var AccountTypeNames = map[xdr.AccountType]string{
	xdr.AccountTypeAccountAnonymousUser:     "anonymous_user",
	xdr.AccountTypeAccountRegisteredUser:    "registered_user",
	xdr.AccountTypeAccountMerchant:          "merchant",
	xdr.AccountTypeAccountDistributionAgent: "distribution_agent",
	xdr.AccountTypeAccountSettlementAgent:   "settlement_agent",
	xdr.AccountTypeAccountExchangeAgent:     "exchange_agent",
	xdr.AccountTypeAccountBank:              "bank",
	xdr.AccountTypeAccountScratchCard:		 "scratch_card",
	xdr.AccountTypeAccountCommission:      	 "commission",
}

func PopulateAccountType(accountType xdr.AccountType) (typeI int32, typ string) {
	var ok bool
	typeI = int32(accountType)
	typ, ok = AccountTypeNames[accountType]
	if !ok {
		typ = "unknown"
	}
	return
}

func PopulateAccountTypeP(accountType xdr.AccountType) (*int32, *string) {
	accTypeI, accType := PopulateAccountType(accountType)
	return &accTypeI, &accType
}

// Account is the summary of an account
type Account struct {
	Links struct {
		Self         hal.Link `json:"self"`
		Transactions hal.Link `json:"transactions"`
		Operations   hal.Link `json:"operations"`
		Payments     hal.Link `json:"payments"`
		Effects      hal.Link `json:"effects"`
		Offers       hal.Link `json:"offers"`
	} `json:"_links"`

	HistoryAccount
	Sequence             string            `json:"sequence"`
	Type                 string            `json:"type"`
	TypeI                int32             `json:"type_i"`
	SubentryCount        int32             `json:"subentry_count"`
	InflationDestination string            `json:"inflation_destination,omitempty"`
	HomeDomain           string            `json:"home_domain,omitempty"`
	Thresholds           AccountThresholds `json:"thresholds"`
	Flags                AccountFlags      `json:"flags"`
	Balances             []Balance         `json:"balances"`
	Signers              []Signer          `json:"signers"`
	Data                 map[string]string `json:"data"`
}

type AccountBalance struct {
	AccountId string `json:"account_id"`
	Balance   string `json:"balance"`
	Limit     string `json:"limit,omitempty"`
}

type MultiAccountAssetBalances struct {
	Asset    details.Asset    `json:"asset"`
	Balances []AccountBalance `json:"balances"`
}
type MultiAssetBalances struct {
	Assets []MultiAccountAssetBalances `json:"assets"`
}

// AccountStatistics is the detailed income/outcome statistics of an account
type AccountStatistics struct {
	Links struct {
		Self    hal.Link `json:"self"`
		Account hal.Link `json:"account"`
	} `json:"_links"`
	Statistics []AccountStatisticsEntry `json:"statistics"`
}

// AccountStatisticsEntry represents account_statistics row
type AccountStatisticsEntry struct {
	AssetType            string `json:"asset_type"`
	AssetCode            string `json:"asset_code"`
	CounterpartyType     int16  `json:"counterparty_type"`
	CounterpartyTypeName string `json:"counterparty_type_name"`

	Income struct {
		Daily   string `json:"daily"`
		Weekly  string `json:"weekly"`
		Monthly string `json:"monthly"`
		Annual  string `json:"annual"`
	} `json:"income"`
	Outcome struct {
		Daily   string `json:"daily"`
		Weekly  string `json:"weekly"`
		Monthly string `json:"monthly"`
		Annual  string `json:"annual"`
	} `json:"outcome"`
}

// AccountFlags represents the state of an account's flags
type AccountFlags struct {
	AuthRequired  bool `json:"auth_required"`
	AuthRevocable bool `json:"auth_revocable"`
}

// AccountThresholds represents an accounts "thresholds", the numerical values
// needed to satisfy the authorization of a given operation.
type AccountThresholds struct {
	LowThreshold  byte `json:"low_threshold"`
	MedThreshold  byte `json:"med_threshold"`
	HighThreshold byte `json:"high_threshold"`
}

// Asset represents a single asset
type Asset details.Asset

type HistoryAsset struct {
	Asset
	ID          int64 `json:"id"`
	IsAnonymous bool  `json:"is_anonymous"`
}

// Balance represents an account's holdings for a single currency type
type Balance struct {
	Balance string `json:"balance"`
	Limit   string `json:"limit,omitempty"`
	details.Asset
}

// HistoryAccount is a simple resource, used for the account collection actions.
// It provides only the "TotalOrderID" of the account and its account id.
type HistoryAccount struct {
	ID        string `json:"id"`
	PT        string `json:"paging_token"`
	AccountID string `json:"account_id"`
}

// Ledger represents a single closed ledger
type Ledger struct {
	Links struct {
		Self         hal.Link `json:"self"`
		Transactions hal.Link `json:"transactions"`
		Operations   hal.Link `json:"operations"`
		Payments     hal.Link `json:"payments"`
		Effects      hal.Link `json:"effects"`
	} `json:"_links"`
	ID               string    `json:"id"`
	PT               string    `json:"paging_token"`
	Hash             string    `json:"hash"`
	PrevHash         string    `json:"prev_hash,omitempty"`
	Sequence         uint32    `json:"sequence"`
	TransactionCount uint32    `json:"transaction_count"`
	OperationCount   uint32    `json:"operation_count"`
	ClosedAt         time.Time `json:"closed_at"`
	TotalCoins       string    `json:"total_coins"`
	FeePool          string    `json:"fee_pool"`
	BaseFee          uint32    `json:"base_fee"`
	BaseReserve      string    `json:"base_reserve"`
	MaxTxSetSize     uint32    `json:"max_tx_set_size"`
}

// Offer is the display form of an offer to trade currency.
type Offer struct {
	Links struct {
		Self       hal.Link `json:"self"`
		OfferMaker hal.Link `json:"offer_maker"`
	} `json:"_links"`

	ID      int64  `json:"id"`
	PT      string `json:"paging_token"`
	Seller  string `json:"seller"`
	Selling Asset  `json:"selling"`
	Buying  Asset  `json:"buying"`
	Amount  string `json:"amount"`
	PriceR  Price  `json:"price_r"`
	Price   string `json:"price"`
}

// OrderBookSummary represents a snapshot summary of a given order book
type OrderBookSummary struct {
	Bids    []PriceLevel `json:"bids"`
	Asks    []PriceLevel `json:"asks"`
	Selling Asset        `json:"base"`
	Buying  Asset        `json:"counter"`
}

// Path represents a single payment path.
type Path struct {
	SourceAssetType        string  `json:"source_asset_type"`
	SourceAssetCode        string  `json:"source_asset_code,omitempty"`
	SourceAssetIssuer      string  `json:"source_asset_issuer,omitempty"`
	SourceAmount           string  `json:"source_amount"`
	DestinationAssetType   string  `json:"destination_asset_type"`
	DestinationAssetCode   string  `json:"destination_asset_code,omitempty"`
	DestinationAssetIssuer string  `json:"destination_asset_issuer,omitempty"`
	DestinationAmount      string  `json:"destination_amount"`
	Path                   []Asset `json:"path"`
}

// Price represents a price
type Price base.Price

// PriceLevel represents an aggregation of offers that share a given price
type PriceLevel struct {
	PriceR Price  `json:"price_r"`
	Price  string `json:"price"`
	Amount string `json:"amount"`
}

// Root is the initial map of links into the api.
type Root struct {
	Links struct {
		Account             hal.Link `json:"account"`
		AccountTransactions hal.Link `json:"account_transactions"`
		Friendbot           hal.Link `json:"friendbot"`
		Metrics             hal.Link `json:"metrics"`
		OrderBook           hal.Link `json:"order_book"`
		Self                hal.Link `json:"self"`
		Transaction         hal.Link `json:"transaction"`
		Transactions        hal.Link `json:"transactions"`
	} `json:"_links"`

	HorizonVersion      string `json:"horizon_version"`
	StellarCoreVersion  string `json:"core_version"`
	HorizonSequence     int32  `json:"horizon_latest_ledger"`
	StellarCoreSequence int32  `json:"core_latest_ledger"`
	NetworkPassphrase   string `json:"network_passphrase"`
}

// Signer represents one of an account's signers.
type Signer struct {
	PublicKey  string `json:"public_key"`
	Weight     int32  `json:"weight"`
	SignerType uint32 `json:"signertype"`
}

// Trade represents a trade effect
type Trade struct {
	Links struct {
		Self   hal.Link `json:"self"`
		Seller hal.Link `json:"seller"`
		Buyer  hal.Link `json:"buyer"`
	} `json:"_links"`

	ID                string `json:"id"`
	PT                string `json:"paging_token"`
	Seller            string `json:"seller"`
	SoldAssetType     string `json:"sold_asset_type"`
	SoldAssetCode     string `json:"sold_asset_code,omitempty"`
	SoldAssetIssuer   string `json:"sold_asset_issuer,omitempty"`
	Buyer             string `json:"buyer"`
	BoughtAssetType   string `json:"bought_asset_type"`
	BoughtAssetCode   string `json:"bought_asset_code,omitempty"`
	BoughtAssetIssuer string `json:"bought_asset_issuer,omitempty"`
}

// Transaction represents a single, successful transaction
type Transaction struct {
	Links struct {
		Self       hal.Link `json:"self"`
		Account    hal.Link `json:"account"`
		Ledger     hal.Link `json:"ledger"`
		Operations hal.Link `json:"operations"`
		Effects    hal.Link `json:"effects"`
		Precedes   hal.Link `json:"precedes"`
		Succeeds   hal.Link `json:"succeeds"`
	} `json:"_links"`
	ID              string    `json:"id"`
	PT              string    `json:"paging_token"`
	Hash            string    `json:"hash"`
	Ledger          int32     `json:"ledger"`
	LedgerCloseTime time.Time `json:"created_at"`
	Account         string    `json:"source_account"`
	AccountSequence string    `json:"source_account_sequence"`
	FeePaid         int32     `json:"fee_paid"`
	OperationCount  int32     `json:"operation_count"`
	EnvelopeXdr     string    `json:"envelope_xdr"`
	ResultXdr       string    `json:"result_xdr"`
	ResultMetaXdr   string    `json:"result_meta_xdr"`
	FeeMetaXdr      string    `json:"fee_meta_xdr"`
	MemoType        string    `json:"memo_type"`
	Memo            string    `json:"memo,omitempty"`
	Signatures      []string  `json:"signatures"`
	ValidAfter      string    `json:"valid_after,omitempty"`
	ValidBefore     string    `json:"valid_before,omitempty"`
}

// TransactionResultCodes represent a summary of result codes returned from
// a single xdr TransactionResult
type TransactionResultCodes struct {
	TransactionCode string   `json:"transaction"`
	OperationCodes  []string `json:"operations,omitempty"`
}

// TransactionSuccess represents the result of a successful transaction
// submission.
type TransactionSuccess struct {
	Links struct {
		Transaction hal.Link `json:"transaction"`
	} `json:"_links"`
	Hash   string `json:"hash"`
	Ledger int32  `json:"ledger"`
	Env    string `json:"envelope_xdr"`
	Result string `json:"result_xdr"`
	Meta   string `json:"result_meta_xdr"`
}

// AccountLimits is the limits set on an account
type AccountLimits struct {
	Links struct {
		Self    hal.Link `json:"self"`
		Account hal.Link `json:"account"`
	} `json:"_links"`
	Account string               `json:"account"`
	Limits  []AccountLimitsEntry `json:"limits"`
}

// AccountLimitsEntry represents limits on a specific currency
type AccountLimitsEntry struct {
	AssetCode       string `json:"asset_code"`
	MaxOperationOut string `json:"max_operation_out"`
	DailyMaxOut     string `json:"daily_max_out"`
	MonthlyMaxOut   string `json:"monthly_max_out"`
	MaxOperationIn  string `json:"max_operation_in"`
	DailyMaxIn      string `json:"daily_max_in"`
	MonthlyMaxIn    string `json:"monthly_max_in"`
}

type Commission struct {
	Id               int64          `json:"id"`
	From             *string        `json:"from,omitempty"`
	To               *string        `json:"to,omitempty"`
	FromAccountType  *string        `json:"from_account_type,omitempty"`
	FromAccountTypeI *int32         `json:"from_account_type_i,omitempty"`
	ToAccountType    *string        `json:"to_account_type,omitempty"`
	ToAccountTypeI   *int32         `json:"to_account_type_i,omitempty"`
	Asset            *details.Asset `json:"asset,omitempty"`
	FlatFee          string         `json:"flat_fee"`
	PercentFee       string         `json:"percent_fee"`
	Weight           int            `json:"weight"`
}

// NewEffect returns a resource of the appropriate sub-type for the provided
// effect record.
func NewEffect(
	ctx context.Context,
	row history.Effect,
) (result hal.Pageable, err error) {
	return effects.New(ctx, row)
}

// NewOperation returns a resource of the appropriate sub-type for the provided
// operation record.
func NewOperation(
	ctx context.Context,
	row history.Operation,
) (result hal.Pageable, err error) {
	return operations.New(ctx, row)
}
