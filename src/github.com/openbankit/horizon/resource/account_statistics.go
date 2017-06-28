package resource

import (
	"fmt"

	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/assets"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/httpx"
	"github.com/openbankit/horizon/render/hal"

	"golang.org/x/net/context"
)

// Populate fills out the resource's fields
func (as *AccountStatistics) Populate(
	ctx context.Context,
	statistics []history.AccountStatistics,
	ha history.Account,
) (err error) {
	// Populate statistics
	as.Statistics = make([]AccountStatisticsEntry, len(statistics))
	for i, stat := range statistics {
		as.Statistics[i].Populate(stat)
	}
	// Construct links
	lb := hal.LinkBuilder{httpx.BaseURL(ctx)}
	accountLink := fmt.Sprintf("/accounts/%s", ha.Address)
	self := fmt.Sprintf("/accounts/%s/statistics", ha.Address)
	as.Links.Self = lb.Link(self)
	as.Links.Account = lb.Link(accountLink)

	return
}

// Populate fills out the resource's fields
func (entry *AccountStatisticsEntry) Populate(stats history.AccountStatistics) {
	// Set asset
	entry.AssetCode = stats.AssetCode
	if len(entry.AssetCode) <= 4 {
		entry.AssetType, _ = assets.String(xdr.AssetTypeAssetTypeCreditAlphanum4)
	} else {
		entry.AssetType, _ = assets.String(xdr.AssetTypeAssetTypeCreditAlphanum12)
	}

	// Set counterparty type
	entry.CounterpartyType = stats.CounterpartyType
	entry.CounterpartyTypeName = xdr.AccountType(stats.CounterpartyType).String()

	// Populate income
	entry.Income.Daily = amount.String(xdr.Int64(stats.DailyIncome))
	entry.Income.Weekly = amount.String(xdr.Int64(stats.WeeklyIncome))
	entry.Income.Monthly = amount.String(xdr.Int64(stats.MonthlyIncome))
	entry.Income.Annual = amount.String(xdr.Int64(stats.AnnualIncome))
	// Populate outcome
	entry.Outcome.Daily = amount.String(xdr.Int64(stats.DailyOutcome))
	entry.Outcome.Weekly = amount.String(xdr.Int64(stats.WeeklyOutcome))
	entry.Outcome.Monthly = amount.String(xdr.Int64(stats.MonthlyOutcome))
	entry.Outcome.Annual = amount.String(xdr.Int64(stats.AnnualOutcome))
}
