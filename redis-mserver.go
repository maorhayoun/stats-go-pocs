package main

// todo: pipeline
// todo: configuration format

import (
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
	"strconv"
	"time"
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
	m.Use(render.Renderer())

	m.Post("/:key", func(pool *redis.Pool, params martini.Params, req *http.Request, r render.Render) {
		key := params["key"]
		value := req.URL.Query().Get("value")

		c := pool.Get()
		defer c.Close()

		tkey := time.Now().Truncate(time.Hour).Format("200601021504")

		_, err := c.Do("HSET", key, "value", value)
		if err != nil {
			r.JSON(400, map[string]interface{}{"error": err})
			return
		}

		if _, err := c.Do("HINCRBY", key, tkey, 1); err != nil {
			r.JSON(400, map[string]interface{}{"error": err})
			return
		}

		r.JSON(200, map[string]interface{}{})
	})

	m.Get("/:key/recent", func(pool *redis.Pool, params martini.Params, req *http.Request, r render.Render) {
		key := params["key"]

		// todo: time frame query term. e.g. 7d, 12h
		hours, err := strconv.Atoi(req.URL.Query().Get("hours"))
		if err != nil {
			days, _ := strconv.Atoi(req.URL.Query().Get("days"))
			hours = 24 * days
		}

		// set query end time. if not by provided, ends at current time
		date, err := time.Parse(time.RFC3339, req.URL.Query().Get("since"))
		if err != nil {
			date = time.Now()
		}

		// in case of begining of hour - include additional hour
		if date == date.Truncate(time.Hour) {
			hours++
		}

		// gather all time keys to query
		var keys []interface{}
		keys = append(keys, key)
		for i := 0; i < hours; i++ {
			tkey := date.Truncate(time.Hour).Format("200601021504")
			date = date.Add(-time.Hour)
			keys = append(keys, tkey)
		}

		// query redis
		c := pool.Get()
		defer c.Close()
		values, err := redis.Values(c.Do("HMGET", keys...))
		if err != nil {
			r.JSON(400, map[string]interface{}{"error": err})
			return
		}

		// scan the []interface{} slice into a []int slice
		var ints []int
		if err = redis.ScanSlice(values, &ints); err != nil {
			r.JSON(400, map[string]interface{}{"error": err})
			return
		}

		total := 0
		for _, num := range ints {
			total += num
		}

		r.JSON(200, map[string]interface{}{"value": total})
	})

	m.Get("/:key", func(pool *redis.Pool, params martini.Params, r render.Render) {
		key := params["key"]

		c := pool.Get()
		defer c.Close()

		value, err := redis.String(c.Do("HGET", key, "value"))

		if err != nil {
			r.JSON(400, map[string]interface{}{"error": err.Error()})
			return
		}
		fmt.Println(err)
		r.JSON(200, map[string]interface{}{"value": value})

	})

	m.Run()

}
