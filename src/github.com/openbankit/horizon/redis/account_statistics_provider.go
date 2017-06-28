package redis

import (
	"github.com/openbankit/go-base/xdr"
	"time"
	"github.com/garyburd/redigo/redis"
)

type AccountStatisticsProviderInterface interface {
	Insert(stats *AccountStatistics, timeout time.Duration) (error)
	Get(accountId, assetCode string, counterparties []xdr.AccountType) (*AccountStatistics, error)
}

type AccountStatisticsProvider struct {
	conn ConnectionInterface
}

func NewAccountStatisticsProvider(conn ConnectionInterface) *AccountStatisticsProvider {
	return &AccountStatisticsProvider{
		conn: conn,
	}
}

// Inserts account's statistics into redis and sets expiration time.
func (c *AccountStatisticsProvider) Insert(stats *AccountStatistics, timeout time.Duration) (error) {
	params := stats.ToArray()
	err := c.conn.HMSet(params...)
	if err != nil {
		return err
	}

	key := stats.GetKey()
	_, err = c.conn.Expire(key, timeout)
	return err
}

// Tries to get account's stats map from redis. Returns nil, if does not exist
func (c *AccountStatisticsProvider) Get(accountId, assetCode string, counterparties []xdr.AccountType) (*AccountStatistics, error) {
	key := GetAccountStatisticsKey(accountId, assetCode)
	data, err := redis.Int64Map(c.conn.HGetAll(key))
	if err != nil || len(data) == 0 {
		return nil, err
	}

	return ReadAccountStatistics(accountId, assetCode, data, counterparties), nil
}
