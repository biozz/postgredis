package main

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func main() {
	// asuming postgredis is started
	c, err := redis.DialURL("redis://localhost:6380/?foo=bar")
	if err != nil {
		// handle connection error
	}
	defer c.Close()
	_, _ = redis.Bool(c.Do("SET", "x", "123"))
	result, _ := redis.Bytes(c.Do("GET", "x"))
	fmt.Println(string(result))
}
