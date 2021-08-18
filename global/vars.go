package global

import (
	"github.com/gomodule/redigo/redis"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var ConnPool *redis.Pool
var ElasticClient *elastic.Client

func RedisInit() {
	ConnPool = newPool(":6379")
	cleanupHook()
}

func ElasticInit() {
	var err error
	ElasticClient, err = elastic.NewClient(
		elastic.SetSniff(true),
		elastic.SetURL("http://localhost:9200"),
		elastic.SetHealthcheckInterval(5*time.Second), // quit trying after 5 seconds
	)
	if err != nil {
		log.Println(err)
		panic(err)
	}

}

func newPool(server string) *redis.Pool {

	return &redis.Pool{

		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func cleanupHook() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		ConnPool.Close()
		os.Exit(0)
	}()
}

func DeleteAllRedis()  {
	conn := ConnPool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHALL")
	if err != nil {
		panic(err)
	}
}
