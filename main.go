package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/view"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

func getRedisHostAndPort() string {
	redis_host := os.Getenv("REDIS_PORT_6379_TCP_ADDR")
	redis_port := os.Getenv("REDIS_PORT_6379_TCP_PORT")
	if len(redis_host) == 0 {
		redis_host = "192.168.118.174"
	}
	if len(redis_port) == 0 {
		redis_port = "6379"
	}
	redis_info := fmt.Sprintf("%s:%s", redis_host, redis_port)
	log.Printf("Redis连接信息: %s", redis_info)
	return redis_info
}

var redis_info = getRedisHostAndPort()

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   3,
		MaxActive: 1000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redis_info)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

var pool = newPool()

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type Link struct {
	Url   string `redis:"url"`
	Title string `redis:"title"`
	Count int    `redis:"count"`
	Key   string
}


func main() {
	app := iris.New()
	app.Adapt(httprouter.New())
	app.Adapt(view.HTML("./templates", ".html"))
	app.Get("/links", func(s *iris.Context) {
		c := pool.Get()
		defer c.Close()
		urlKyes, err := redis.Values(c.Do("keys", "short-url:*"))

		if err != nil {
			panic(err)
		}
		var links []Link
		leng:=len(urlKyes)
		for i := 0; i < leng; i++ {
			key := string(urlKyes[leng-i-1].([]byte))
			value, err := redis.Values(c.Do("HGETALL", key))
			if err != nil {
				panic(err)
			}
			var link Link
			if err := redis.ScanStruct(value, &link); err != nil {
				fmt.Println(err)
			}
			linkKey := strings.Split(key, ":")[1]
			link.Key = fmt.Sprint(s.ServerHost(), "/r/", linkKey)
			links = append(links, link)
		}
		s.Render("links/index.html", links)
	})

	app.Post("/links", func(s *iris.Context) {
		url := s.FormValue("link_url")
		title := s.FormValue("link_title")
		var link Link

		link.Url = string(url)
		link.Title = string(title)
		link.Count = 0

		c := pool.Get()
		defer c.Close()

		key := fmt.Sprint("short-url:", RandStringRunes(7))

		if _, err := c.Do("HMSET", redis.Args{}.Add(key).AddFlat(&link)...); err != nil {
			panic(err)
		}

		s.Redirect("/links")
	})

	app.Get("/r/:key", func(s *iris.Context) {
		queryKey := s.Param("key")
		key := fmt.Sprint("short-url:", queryKey)
		c := pool.Get()
		defer c.Close()

		url, err := redis.String(c.Do("HGET", key, "url"))

		if err != nil {
			s.Writef("not found")
		} else {
			c.Do("HINCRBY", key, "count", 1)
			s.Redirect(url)
		}
	})

	app.Listen(":8001")
}
