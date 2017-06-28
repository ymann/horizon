package redis

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

type ProcessedOpProviderInterface interface {
	Insert(processedOp *ProcessedOp, timeout time.Duration) error
	Get(txHash string, opIndex int, isIncoming bool) (*ProcessedOp, error)
	Delete(txHash string, opIndex int, isIncoming bool) error
}

type ProcessedOpProvider struct {
	conn ConnectionInterface
}

func NewProcessedOpProvider(conn ConnectionInterface) *ProcessedOpProvider {
	return &ProcessedOpProvider{
		conn: conn,
	}
}

// Inserts processed op into redis and sets expiration time
func (c *ProcessedOpProvider) Insert(processedOp *ProcessedOp, timeout time.Duration) error {
	rawData := processedOp.ToArray()
	err := c.conn.HMSet(rawData...)
	if err != nil {
		return err
	}

	key := processedOp.GetKey()
	_, err = c.conn.Expire(key, timeout)
	return err
}

// Tries to get account's stats map from redis. Returns nil, if does not exist
func (c *ProcessedOpProvider) Get(txHash string, opIndex int, isIncoming bool) (*ProcessedOp, error) {
	key := GetProcessedOpKey(txHash, opIndex, isIncoming)
	data, err := redis.Int64Map(c.conn.HGetAll(key))
	if err != nil || len(data) == 0 {
		return nil, err
	}

	return ReadProcessedOp(txHash, opIndex, isIncoming, data), nil
}

func (c *ProcessedOpProvider) Delete(txHash string, opIndex int, isIncoming bool) error {
	key := GetProcessedOpKey(txHash, opIndex, isIncoming)
	return c.conn.Delete(key)
}
