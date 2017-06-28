package resource

import (
	"fmt"

	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/httpx"
	"github.com/openbankit/horizon/render/hal"

	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/xdr"
	"golang.org/x/net/context"
)

// Populate fills out the resource's fields
func (al *AccountLimits) Populate(
	ctx context.Context,
	address string,
	limits []history.AccountLimits,
) (err error) {
	// Populate limits
	al.Account = address
	al.Limits = make([]AccountLimitsEntry, len(limits))
	for i, limit := range limits {
		al.Limits[i].Populate(limit)
	}

	// Construct links
	lb := hal.LinkBuilder{httpx.BaseURL(ctx)}
	accountLink := fmt.Sprintf("/accounts/%s", address)
	self := fmt.Sprintf("/accounts/%s/limits", address)
	al.Links.Self = lb.Link(self)
	al.Links.Account = lb.Link(accountLink)

	return
}

// Populate fills out the resource's fields
func (ale *AccountLimitsEntry) Populate(entry history.AccountLimits) {
	ale.AssetCode = entry.AssetCode
	ale.MaxOperationOut = ale.formatLimit(entry.MaxOperationOut)
	ale.DailyMaxOut = ale.formatLimit(entry.DailyMaxOut)
	ale.MonthlyMaxOut = ale.formatLimit(entry.MonthlyMaxOut)
	ale.MaxOperationIn = ale.formatLimit(entry.MaxOperationIn)
	ale.DailyMaxIn = ale.formatLimit(entry.DailyMaxIn)
	ale.MonthlyMaxIn = ale.formatLimit(entry.MonthlyMaxIn)
}

func (ale *AccountLimitsEntry) formatLimit(limit int64) string {
	if limit == -1 {
		return amount.String(xdr.Int64(limit) * amount.One)
	}
	return amount.String(xdr.Int64(limit))
}
