package cache

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"database/sql"
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

var historyAssetCache *cache.Cache
var historyAssetCacheOnce sync.Once

// HistoryAsset provides a cached lookup of asset values from
// xdr.Asset.
type HistoryAsset struct {
	*cache.Cache
	history history.QInterface
}

func getHistoryAssetCache() *cache.Cache {
	historyAssetCacheOnce.Do(func() {
		historyAssetCache = cache.New(time.Duration(10)*time.Minute, time.Duration(1)*time.Minute)
	})
	return historyAssetCache
}

// NewHistoryAsset initializes a new instance of `HistoryAsset`
func NewHistoryAsset(db history.QInterface) *HistoryAsset {
	return &HistoryAsset{
		Cache:   getHistoryAssetCache(),
		history: db,
	}
}

// Get looks up the history.Asset for the given xdr.Asset.
func (c *HistoryAsset) Get(asset xdr.Asset) (*history.Asset, error) {
	var typ xdr.AssetType
	var issuer, code string
	err := asset.Extract(&typ, &code, &issuer)
	if err != nil {
		return nil, err
	}

	assetKey := issuer + code
	found, ok := c.Cache.Get(assetKey)
	if ok {
		return found.(*history.Asset), nil
	}

	result := new(history.Asset)
	err = c.history.AssetByParams(result, int(typ), code, issuer)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		result = nil
	}
	c.Cache.Set(assetKey, result, cache.DefaultExpiration)
	return result, nil
}
