package cache

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/dgraph-io/badger"
	"github.com/gomodule/redigo/redis"
	"log"
	"os"
	"testing"
	"time"
)

var testRedisCache RedisCache
var testBadgerCache BadgerCache

func TestMain(m *testing.M) {
	// setting up miniredis server
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	pool := redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", s.Addr())
		},
		TestOnBorrow: nil,
		MaxIdle:      50,
		MaxActive:    10000,
		IdleTimeout:  5 * time.Minute,
	}
	// populate the instance of testRedisCache
	testRedisCache.Conn = &pool
	testRedisCache.Prefix = "test-gudu"

	defer func(Conn *redis.Pool) {
		_ = Conn.Close()
	}(testRedisCache.Conn)

	_ = os.RemoveAll("./testdata/tmp/badger")
	if _, err := os.Stat("./testdata/tmp"); os.IsNotExist(err) {
		err := os.Mkdir("testdata/tmp", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err = os.Mkdir("./testdata/tmp/badger", 0755); err != nil {
		log.Fatal(err)
	}

	db, _ := badger.Open(badger.DefaultOptions("./testdata/tmp/badger"))
	testBadgerCache.Conn = db
	testBadgerCache.Prefix = "test-gudu"

	os.Exit(m.Run())
}
