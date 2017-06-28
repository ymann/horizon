// Package cache provides various caches used in horizon.
package cache

import (
	"github.com/openbankit/horizon/db2/history"
	"database/sql"
	"github.com/patrickmn/go-cache"
	"time"
)

// HistoryAccount provides a cached lookup of history_account_id values from
// account addresses.
type HistoryAccount struct {
	*cache.Cache
	q history.QInterface
}

// NewHistoryAccount initializes a new instance of `HistoryAccount`
func NewHistoryAccount(historyQ history.QInterface) *HistoryAccount {
	return NewHistoryAccountWithExp(historyQ, cache.NoExpiration, cache.NoExpiration)
}

func NewHistoryAccountWithExp(historyQ history.QInterface, defaultExpiration, cleanupInterval time.Duration) *HistoryAccount {
	return &HistoryAccount{
		Cache: cache.New(defaultExpiration, cleanupInterval),
		q:     historyQ,
	}
}

// Get looks up the History Account ID (i.e. the ID of the operation that
// created the account) for the given strkey encoded address. Returns sql.ErrNoRows
// if account does not exists
func (c *HistoryAccount) Get(address string) (*history.Account, error) {
	found, ok := c.Cache.Get(address)

	if ok {
		if found == nil {
			return nil, sql.ErrNoRows
		}
		result := found.(*history.Account)
		if result == nil {
			return nil, sql.ErrNoRows
		}
		return result, nil
	}

	var rawResult history.Account
	err := c.q.AccountByAddress(&rawResult, address)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Add(address, nil)
		}
		return nil, err
	}

	result := &rawResult
	c.Add(address, result)
	return result, nil
}

// Adds address-id pair into cache
func (c *HistoryAccount) Add(address string, account *history.Account) {
	c.Cache.Set(address, account, cache.DefaultExpiration)
}
