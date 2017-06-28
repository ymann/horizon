package helpers

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/go-base/amount"
)

// assetDetails sets the details for `a` on `result` using keys with `prefix`
func AssetDetails(result map[string]interface{}, a xdr.Asset, prefix string) error {
	var (
		t    string
		code string
		i    string
	)
	err := a.Extract(&t, &code, &i)
	if err != nil {
		return err
	}
	result[prefix+"asset_type"] = t

	if a.Type == xdr.AssetTypeAssetTypeNative {
		return nil
	}

	result[prefix+"asset_code"] = code
	result[prefix+"asset_issuer"] = i
	return nil
}

func FlagDetails(flagDetails map[string]bool, flagPtr *xdr.Uint32, setValue bool) {
	if flagPtr != nil {
		flags := xdr.AccountFlags(*flagPtr)

		if flags&xdr.AccountFlagsAuthRequiredFlag != 0 {
			flagDetails["auth_required_flag"] = setValue
		}
		if flags&xdr.AccountFlagsAuthRevocableFlag != 0 {
			flagDetails["auth_revocable_flag"] = setValue
		}
		if flags&xdr.AccountFlagsAuthImmutableFlag != 0 {
			flagDetails["auth_immutable_flag"] = setValue
		}
	}
}

func TradeDetails(buyer, seller xdr.AccountId, claim xdr.ClaimOfferAtom) (bd map[string]interface{}, sd map[string]interface{}) {
	bd = map[string]interface{}{
		"offer_id":      claim.OfferId,
		"seller":        seller.Address(),
		"bought_amount": amount.String(claim.AmountSold),
		"sold_amount":   amount.String(claim.AmountBought),
	}
	AssetDetails(bd, claim.AssetSold, "bought_")
	AssetDetails(bd, claim.AssetBought, "sold_")

	sd = map[string]interface{}{
		"offer_id":      claim.OfferId,
		"seller":        buyer.Address(),
		"bought_amount": amount.String(claim.AmountBought),
		"sold_amount":   amount.String(claim.AmountSold),
	}
	AssetDetails(sd, claim.AssetBought, "bought_")
	AssetDetails(sd, claim.AssetSold, "sold_")

	return
}
