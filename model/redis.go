package model

import "github.com/garyburd/redigo/redis"
import "log"

var conn redis.Conn

// Sets up the global Redis connection
func init() {
	var err error
	conn, err = redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}
}
