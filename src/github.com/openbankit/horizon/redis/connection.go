package redis

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

const (
	request_state_queued = "QUEUED"
)

type ConnectionInterface interface {

	// Sets the specified fields to their respective values in the hash stored at key.
	// This command overwrites any existing fields in the hash.
	// If key does not exist, a new key holding a hash is created.
	HMSet(args ...interface{}) error

	// Returns all fields and values of the hash stored at key. In the returned value,
	// every field name is followed by its value, so the length of the reply is twice the size of the hash.
	HGetAll(key string) (interface{}, error)

	// Set a timeout on key. After the timeout has expired, the key will automatically be deleted.
	Expire(key string, timeout time.Duration) (bool, error)

	// Atomically sets key to value and returns the old value stored at key.
	GetSet(key string, data interface{}) (interface{}, error)

	// Get the value of key. If the key does not exist the special value nil is returned.
	// An error is returned if the value stored at key is not a string, because GET only handles string values.
	Get(key string) (interface{}, error)

	// Set key to hold the string value. If key already holds a value, it is overwritten, regardless of its type.
	// Any previous time to live associated with the key is discarded on successful SET operation.
	Set(key string, data interface{}) error

	// Marks the given keys to be watched for conditional execution of a transaction.
	Watch(key string) error

	// Flushes all the previously watched keys for a transaction. If you call EXEC or DISCARD, there's no need to manually call UNWATCH.
	UnWatch() error

	// Marks the start of a transaction block. Subsequent commands will be queued for atomic execution using EXEC.
	Multi() error

	// Executes all previously queued commands in a transaction and restores the connection state to normal.
	Exec() (bool, error)

	// Close closes the connection.
	Close() error

	// Removes the specified keys. A key is ignored if it does not exist.
	Delete(key string) error

	Ping() error
}

type Connection struct {
	redis.Conn
}

func NewConnection(c redis.Conn) *Connection {
	return &Connection{
		Conn: c,
	}
}

// Sets the specified fields to their respective values in the hash stored at key.
// This command overwrites any existing fields in the hash.
// If key does not exist, a new key holding a hash is created.
func (r *Connection) HMSet(args ...interface{}) error {
	_, err := r.Do("HMSET", args...)
	return err
}

// Returns all fields and values of the hash stored at key. In the returned value,
// every field name is followed by its value, so the length of the reply is twice the size of the hash.
func (r *Connection) HGetAll(key string) (interface{}, error) {
	return r.Do("HGETALL", key)
}

// Set a timeout on key. After the timeout has expired, the key will automatically be deleted.
func (r *Connection) Expire(key string, timeout time.Duration) (bool, error) {
	timeoutInSeconds := int64(timeout / time.Second)
	resp, err := r.Do("EXPIRE", key, timeoutInSeconds)
	if resp == request_state_queued {
		return true, err
	}

	isSet, err := redis.Bool(resp, err)
	return isSet, err
}

// Atomically sets key to value and returns the old value stored at key.
func (r *Connection) GetSet(key string, data interface{}) (interface{}, error) {
	return r.Do("GETSET", key, data)
}

// Get the value of key. If the key does not exist the special value nil is returned.
// An error is returned if the value stored at key is not a string, because GET only handles string values.
func (r *Connection) Get(key string) (interface{}, error) {
	return r.Do("GET", key)
}

// Set key to hold the string value. If key already holds a value, it is overwritten, regardless of its type.
// Any previous time to live associated with the key is discarded on successful SET operation.
func (r *Connection) Set(key string, data interface{}) error {
	_, err := r.Do("SET", key, data)
	return err
}

// Marks the given keys to be watched for conditional execution of a transaction.
func (r *Connection) Watch(key string) error {
	_, err := r.Do("WATCH", key)
	return err
}

// Flushes all the previously watched keys for a transaction. If you call EXEC or DISCARD, there's no need to manually call UNWATCH.
func (r *Connection) UnWatch() error {
	_, err := r.Do("UNWATCH")
	return err
}

// Marks the start of a transaction block. Subsequent commands will be queued for atomic execution using EXEC.
func (r *Connection) Multi() error {
	_, err := r.Do("MULTI")
	return err
}

// Executes all previously queued commands in a transaction and restores the connection state to normal.
func (r *Connection) Exec() (bool, error) {
	resp, err := r.Do("EXEC")
	if resp == nil || err != nil {
		return false, err
	}
	return true, nil
}

// Removes the specified keys. A key is ignored if it does not exist.
func (r *Connection) Delete(key string) error {
	_, err := r.Do("DEL", key)
	return err
}

func (r *Connection) Ping() error {
	_, err := r.Do("PING")
	return err
}

func IsConnectionClosed(err error) bool {
	return err.Error() == "redigo: connection closed"
}
