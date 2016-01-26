package model

import "github.com/garyburd/redigo/redis"
import "time"

var pool *redis.Pool

// Sets up the global Redis connection
func init() {
	pool = &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", ":6379")
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
}
