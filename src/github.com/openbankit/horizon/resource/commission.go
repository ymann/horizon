package resource

import (
	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/db2/history/details"
	"fmt"
)

// Populate fills out the Commission
func (res *Commission) Populate(row history.Commission) (err error) {
	key := row.GetKey()
	res.Weight = key.CountWeight()
	res.Id = row.ID

	if key.From != "" {
		res.From = &key.From
	}

	if key.To != "" {
		res.To = &key.To
	}
	if key.FromType != nil {
		res.FromAccountTypeI, res.FromAccountType = PopulateAccountTypeP(xdr.AccountType(*key.FromType))
	}
	if key.ToType != nil {
		res.ToAccountTypeI, res.ToAccountType = PopulateAccountTypeP(xdr.AccountType(*key.ToType))
	}

	if (key.Asset != details.Asset{}) {
		res.Asset = &key.Asset
	}
	res.FlatFee = amount.String(xdr.Int64(row.FlatFee))
	res.PercentFee = amount.String(xdr.Int64(row.PercentFee))
	return
}

func (this Commission) PagingToken() string {
	return fmt.Sprintf("%d", this.Id)
}
