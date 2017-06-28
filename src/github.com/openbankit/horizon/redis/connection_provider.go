package redis

import "github.com/openbankit/horizon/log"

const max_connection_reties = 10

type ConnectionProviderInterface interface {
	GetConnection() ConnectionInterface
}

type ConnectionProvider struct {
}

func NewConnectionProvider() ConnectionProviderInterface {
	return &ConnectionProvider{}
}

func (c ConnectionProvider) GetConnection() ConnectionInterface {
	if redisPool == nil {
		log.Panic("Redis must be initialized")
	}
	for i := 0; i < max_connection_reties; i++ {
		conn := redisPool.Get()
		if conn.Err() != nil {
			conn.Close()
			continue
		}
		return NewConnection(conn)
	}
	return NewConnection(redisPool.Get())
}
