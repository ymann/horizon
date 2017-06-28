package validators

import (
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/cache"
)

type AssetsValidatorInterface interface {
	GetValidAsset(asset xdr.Asset) (*history.Asset, error)
	IsAssetValid(asset xdr.Asset) (bool, error)
	IsAssetsValid(assets ...xdr.Asset) (bool, error)
}

type AssetsValidator struct {
	assetsProvider *cache.HistoryAsset
}

func NewAssetsValidator(historyQ history.QInterface) *AssetsValidator {
	return &AssetsValidator{
		assetsProvider: cache.NewHistoryAsset(historyQ),
	}
}

func (v *AssetsValidator) GetValidAsset(asset xdr.Asset) (*history.Asset, error) {
	return v.assetsProvider.Get(asset)
}

func (v *AssetsValidator) IsAssetValid(asset xdr.Asset) (bool, error) {
	storedAsset, err := v.GetValidAsset(asset)
	return storedAsset != nil, err
}

func (v *AssetsValidator) IsAssetsValid(assets ...xdr.Asset) (bool, error) {
	for _, asset := range assets {
		isValid, err := v.IsAssetValid(asset)
		if err != nil {
			return false, err
		}

		if !isValid {
			return isValid, err
		}
	}
	return true, nil
}