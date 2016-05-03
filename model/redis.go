package model

import "github.com/garyburd/redigo/redis"
import "time"

var pool *redis.Pool

// Sets up the global Redis connection
func InitPool(db int) *redis.Pool {
	pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", ":6379", redis.DialDatabase(db))
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}

	return pool
}
