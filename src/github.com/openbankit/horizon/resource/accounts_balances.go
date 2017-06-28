package resource

import (
	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/horizon/assets"
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history/details"
	"golang.org/x/net/context"
)

// Populate fills out the resource's fields
func (this *MultiAssetBalances) Populate(
	ctx context.Context,
	ct []core.Trustline,
) (err error) {
	var assetsMap = make(map[details.Asset][]AccountBalance)
	for _, tl := range ct {
		assetType, err := assets.String(tl.Assettype)
		if err != nil {
			return err
		}
		asset := details.Asset{assetType, tl.Assetcode, tl.Issuer}
		accBalance := AccountBalance{tl.Accountid, amount.String(tl.Balance), amount.String(tl.Tlimit)}
		assetsMap[asset] = append(assetsMap[asset], accBalance)
	}

	// populate balances
	this.Assets = make([]MultiAccountAssetBalances, len(assetsMap))
	i := 0
	for key, value := range assetsMap {
		this.Assets[i].Asset = key
		this.Assets[i].Balances = value
		i++
	}
	return
}
