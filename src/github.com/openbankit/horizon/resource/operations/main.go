package operations

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/db2/history/details"
	"github.com/openbankit/horizon/render/hal"
	"github.com/openbankit/horizon/resource/base"
	"golang.org/x/net/context"
	"time"
)

// TypeNames maps from operation type to the string used to represent that type
// in horizon's JSON responses
var TypeNames = map[xdr.OperationType]string{
	xdr.OperationTypeCreateAccount:      "create_account",
	xdr.OperationTypePayment:            "payment",
	xdr.OperationTypePathPayment:        "path_payment",
	xdr.OperationTypeManageOffer:        "manage_offer",
	xdr.OperationTypeCreatePassiveOffer: "create_passive_offer",
	xdr.OperationTypeSetOptions:         "set_options",
	xdr.OperationTypeChangeTrust:        "change_trust",
	xdr.OperationTypeAllowTrust:         "allow_trust",
	xdr.OperationTypeAccountMerge:       "account_merge",
	xdr.OperationTypeInflation:          "inflation",
	xdr.OperationTypeManageData:         "manage_data",
	xdr.OperationTypeAdministrative:     "administrative",
	xdr.OperationTypePaymentReversal:    "payment_reversal",
	xdr.OperationTypeExternalPayment:    "external_payment",
}

// New creates a new operation resource, finding the appropriate type to use
// based upon the row's type.
func New(
	ctx context.Context,
	row history.Operation,
) (result hal.Pageable, err error) {

	base := Base{}
	base.Populate(ctx, row)

	switch row.Type {
	case xdr.OperationTypeCreateAccount:
		e := CreateAccount{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypePayment:
		e := Payment{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypePathPayment:
		e := PathPayment{}
		e.Payment.Base = base
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypePaymentReversal:
		e := PaymentReversal{}
		e.Base = base
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeManageOffer:
		e := ManageOffer{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeCreatePassiveOffer:
		e := CreatePassiveOffer{}
		e.ManageOffer.Base = base
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeSetOptions:
		e := SetOptions{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeChangeTrust:
		e := ChangeTrust{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeAllowTrust:
		e := AllowTrust{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeAccountMerge:
		e := AccountMerge{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeInflation:
		e := Inflation{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeManageData:
		e := ManageData{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeAdministrative:
		e := Administrative{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeExternalPayment:
		e := ExternalPayment{Base: base}
		err = row.UnmarshalDetails(&e)
		result = e
	default:
		result = base
	}

	return
}

type Base struct {
	Links struct {
		Self        hal.Link `json:"self"`
		Transaction hal.Link `json:"transaction"`
		Effects     hal.Link `json:"effects"`
		Succeeds    hal.Link `json:"succeeds"`
		Precedes    hal.Link `json:"precedes"`
	} `json:"_links"`

	ID            string    `json:"id"`
	PT            string    `json:"paging_token"`
	SourceAccount string    `json:"source_account"`
	Type          string    `json:"type"`
	TypeI         int32     `json:"type_i"`
	ClosedAt      time.Time `json:"closed_at"`
}

type CreateAccount struct {
	Base
	Fee         details.Fee `json:"fee"`
	AccountType int32       `json:"account_type"`
	Funder      string      `json:"funder"`
	Account     string      `json:"account"`
}

type Payment struct {
	Base
	details.Payment
}

type PathPayment struct {
	Payment
	Fee               details.Fee     `json:"fee"`
	Path              []details.Asset `json:"path"`
	SourceMax         string          `json:"source_max"`
	SourceAssetType   string          `json:"source_asset_type"`
	SourceAssetCode   string          `json:"source_asset_code,omitempty"`
	SourceAssetIssuer string          `json:"source_asset_issuer,omitempty"`
}

type PaymentReversal struct {
	Base
	PaymentID     int64  `json:"payment_id"`
	PaymentSource string `json:"payment_source"`
	Amount        string `json:"amount"`
	Commission    string `json:"commission"`
	details.Asset
}

// ManageData represents a ManageData operation as it is serialized into json
// for the horizon API.
type ManageData struct {
	Base
	Fee   details.Fee `json:"fee"`
	Name  string      `json:"name"`
	Value string      `json:"value"`
}

type Administrative struct {
	Base
	Details map[string]interface{} `json:"details"`
}

type ManageOffer struct {
	Base
	Fee                details.Fee `json:"fee"`
	OfferID            int64       `json:"offer_id"`
	Amount             string      `json:"amount"`
	Price              string      `json:"price"`
	PriceR             base.Price  `json:"price"`
	BuyingAssetType    string      `json:"buying_asset_type"`
	BuyingAssetCode    string      `json:"buying_asset_code,omitempty"`
	BuyingAssetIssuer  string      `json:"buying_asset_issuer,omitempty"`
	SellingAssetType   string      `json:"selling_asset_type"`
	SellingAssetCode   string      `json:"selling_asset_code,omitempty"`
	SellingAssetIssuer string      `json:"selling_asset_issuer,omitempty"`
}

type CreatePassiveOffer struct {
	ManageOffer
}

type SetOptions struct {
	Base
	Fee           details.Fee `json:"fee"`
	HomeDomain    string      `json:"home_domain,omitempty"`
	InflationDest string      `json:"inflation_dest,omitempty"`

	MasterKeyWeight *int   `json:"master_key_weight,omitempty"`
	SignerKey       string `json:"signer_key,omitempty"`
	SignerWeight    *int   `json:"signer_weight,omitempty"`
	SignerType      *int   `json:"signer_type,omitempty"`

	SetFlags    []int    `json:"set_flags,omitempty"`
	SetFlagsS   []string `json:"set_flags_s,omitempty"`
	ClearFlags  []int    `json:"clear_flags,omitempty"`
	ClearFlagsS []string `json:"clear_flags_s,omitempty"`

	LowThreshold  *int `json:"low_threshold,omitempty"`
	MedThreshold  *int `json:"med_threshold,omitempty"`
	HighThreshold *int `json:"high_threshold,omitempty"`
}

type ChangeTrust struct {
	Base
	details.Asset
	Fee     details.Fee `json:"fee"`
	Limit   string      `json:"limit"`
	Trustee string      `json:"trustee"`
	Trustor string      `json:"trustor"`
}

type AllowTrust struct {
	Base
	details.Asset
	Fee       details.Fee `json:"fee"`
	Trustee   string      `json:"trustee"`
	Trustor   string      `json:"trustor"`
	Authorize bool        `json:"authorize"`
}

type AccountMerge struct {
	Base
	Fee     details.Fee `json:"fee"`
	Account string      `json:"account"`
	Into    string      `json:"into"`
}

type Inflation struct {
	Base
	Fee details.Fee `json:"fee"`
}

type ExternalPayment struct {
	Base
	details.Asset
	Fee    			details.Fee `json:"fee"`
	From   			string	`json:"from"`
	ExchangeAgent		string	`json:"exchangeAgent"`
	DestinationBank		string	`json:"destinationBank"`
	DestinationAccount	string	`json:"destinationAccount"`
	Amount			string	`json:"amount"`
}
