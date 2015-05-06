package main

import (
	"fmt"
	//"log"
	//"time"
	"flag"
	"github.com/garyburd/redigo/redis"
	"github.com/go-martini/martini"
	"net/http"
)

var (
	redisAddress   = flag.String("redis-address", ":6379", "Address to the Redis server")
	maxConnections = flag.Int("max-connections", 10, "Max connections to Redis")
)

func main() {

	flag.Parse()
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			return nil, err
		}

		return c, err
	}, *maxConnections)

	defer redisPool.Close()

	m := martini.Classic()
	m.Map(redisPool)

	// m.Get("/", func() string {
	// 	return "martini Hello world!"
	// })

	m.Post("/:key", func(pool *redis.Pool, params martini.Params, req *http.Request) bool {
		key := params["key"]
		value := req.URL.Query().Get("value")

		c := pool.Get()
		defer c.Close()

		_, err := c.Do("HSET", key, "value", value)
		if err != nil {
			return false
		}

		//exists, err := redis.Bool(c.Do("HEXISTS", key, "counter"))
		//if exists == false {
		//	c.Do("HSET", key, "counter", 0)
		//}

		if _, err := c.Do("HINCRBY", key, "counter", 1); err != nil {
			//log.Fatal(err)
			fmt.Println(err)
			return false
		}

		return true
	})

	m.Get("/counter/:key", func(pool *redis.Pool, params martini.Params) string {
		key := params["key"]

		c := pool.Get()
		defer c.Close()

		value, err := redis.String(c.Do("HGET", key, "counter"))
		fmt.Println(err)
		fmt.Println(value)
		return value
	})

	m.Get("/:key", func(pool *redis.Pool, params martini.Params) string {
		key := params["key"]

		c := pool.Get()
		defer c.Close()

		value, err := redis.String(c.Do("GET", key))

		if err != nil {
			return "error"
		}

		return value
	})

	m.Run()
}
