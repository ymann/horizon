package cache

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/patrickmn/go-cache"
	"strconv"
)

func newAccountStatsKey(address string, assetCode string, counterpartyType xdr.AccountType) string {
	return address + assetCode + strconv.Itoa(int(counterpartyType))
}

// AccountStatistics provides a cached lookup of history_account_id values from
// account addresses.
type AccountStatistics struct {
	*cache.Cache
	q history.QInterface
}

// NewAccountStatistics initializes a new instance of `AccountStatistics`
func NewAccountStatistics(q history.QInterface) *AccountStatistics {
	return &AccountStatistics{
		Cache: cache.New(cache.NoExpiration, cache.NoExpiration),
		q:     q,
	}
}

// Get looks up the history account statistics for the given strkey encoded address, assetCode and counterparty type.
func (c *AccountStatistics) Get(address string, assetCode string, counterPartyType xdr.AccountType) (result *history.AccountStatistics, err error) {
	key := newAccountStatsKey(address, assetCode, counterPartyType)
	found, ok := c.Cache.Get(key)
	if ok {
		result = found.(*history.AccountStatistics)
		return
	}

	stats, err := c.q.GetAccountStatistics(address, assetCode, counterPartyType)
	if err != nil {
		return
	}

	result = &stats
	c.AddWithKey(key, result)
	return
}

func (c *AccountStatistics) AddWithKey(key string, stats *history.AccountStatistics) {
	c.Cache.Set(key, stats, cache.DefaultExpiration)
}

func (c *AccountStatistics) AddWithParams(account, assetCode string, counterparty xdr.AccountType, stats *history.AccountStatistics) {
	key := newAccountStatsKey(account, assetCode, counterparty)
	c.AddWithKey(key, stats)
}

// Adds address-id pair into cache
func (c *AccountStatistics) Add(stats *history.AccountStatistics) {
	key := newAccountStatsKey(stats.Account, stats.AssetCode, xdr.AccountType(stats.CounterpartyType))
	c.AddWithKey(key, stats)
}
