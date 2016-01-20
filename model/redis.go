package model

import "github.com/mediocregopher/radix.v2/redis"
import "log"

var client *redis.Client

func init() {
	var err error
	client, err = redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Panic(err)
	}
}
