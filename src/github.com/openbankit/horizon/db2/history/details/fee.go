package details

import (
	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/xdr"
)

type Fee struct {
	Type          string  `json:"type"`
	TypeI         int32   `json:"type_i"`
	AmountCharged *string `json:"amount_changed,omitempty"`
	FlatFee       *string `json:"flat_fee,omitempty"`
	PercentFee    *string `json:"percent_fee,omitempty"`
}

var FeeTypeNames = map[xdr.OperationFeeType]string{
	xdr.OperationFeeTypeOpFeeNone:    "none",
	xdr.OperationFeeTypeOpFeeCharged: "charged",
}

func (f *Fee) ToMap() (details map[string]interface{}) {
	details = make(map[string]interface{})
	details["type"] = f.Type
	details["type_i"] = f.TypeI
	if f.AmountCharged != nil {
		details["amount_changed"] = *f.AmountCharged
	}

	if f.FlatFee != nil {
		details["flat_fee"] = *f.FlatFee
	}

	if f.PercentFee != nil {
		details["percent_fee"] = *f.PercentFee
	}
	return
}

func (f *Fee) Populate(xdrFee xdr.OperationFee) {
	feeType, feeTypeOk := FeeTypeNames[xdrFee.Type]
	if !feeTypeOk {
		feeType = "unknown"
	}
	f.Type = feeType
	f.TypeI = int32(xdrFee.Type)

	switch xdrFee.Type {
	case xdr.OperationFeeTypeOpFeeCharged:
		charged := xdrFee.MustFee()
		f.AmountCharged = new(string)
		*f.AmountCharged = amount.String(charged.AmountToCharge)

		if charged.FlatFee != nil {
			f.FlatFee = new(string)
			*f.FlatFee = amount.String(*charged.FlatFee)
		}

		if charged.PercentFee != nil {
			f.PercentFee = new(string)
			*f.PercentFee = amount.String(*charged.PercentFee)
		}
	}
}
