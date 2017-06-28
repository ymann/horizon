package redis

import (
	"github.com/garyburd/redigo/redis"
	"errors"
	"time"
	"net/url"
)


var redisPool *redis.Pool

func GetPool() *redis.Pool {
	return redisPool
}

func Init(urlStr string) error {
	if urlStr == "" {
		return errors.New("Redis URL can not be empty!")
	}

	redisURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	redisPool = &redis.Pool{
		MaxIdle:      3,
		IdleTimeout:  240 * time.Second,
		Dial:         dialRedis(redisURL),
		TestOnBorrow: pingRedis,
	}

	// test the connection
	c := redisPool.Get()
	defer c.Close()

	_, err = c.Do("PING")
	if err != nil {
		return err
	}

	return nil
}

func dialRedis(redisURL *url.URL) func() (redis.Conn, error) {
	return func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", redisURL.Host)
		if err != nil {
			return nil, err
		}

		if redisURL.User == nil {
			return c, err
		}

		if pass, ok := redisURL.User.Password(); ok {
			if _, err := c.Do("AUTH", pass); err != nil {
				c.Close()
				return nil, err
			}
		}

		return c, err
	}
}

func pingRedis(c redis.Conn, t time.Time) error {
	_, err := c.Do("PING")
	return err
}
