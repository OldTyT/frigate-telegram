package redis

import (
	"context"
	"time"

	"github.com/oldtyt/frigate-telegram/internal/config"
	"github.com/oldtyt/frigate-telegram/internal/log"
	redis "github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var conf = config.New()
var rdb = redis.NewClient(&redis.Options{
	Addr:     conf.RedisAddr,
	Password: conf.RedisPassword, // no password set
	DB:       conf.RedisDB,       // use default DB
	Protocol: conf.RedisProtocol, // specify 2 for RESP 2 or 3 for RESP 3
})

func AddNewEvent(EventID string, State string) {
	RedisTTL := time.Duration(conf.RedisTTL) * time.Second
	err := rdb.Set(ctx, EventID, State, RedisTTL).Err()
	if err != nil {
		log.Error.Fatalln(err)
	}
}

func CheckEvent(EventID string) bool {
	event, err := rdb.Exists(ctx, EventID).Result()
	if err != nil {
		log.Error.Fatalln(err)
	}
	if event == 0 {
		return true
	}
	val, err := rdb.Get(ctx, EventID).Result()
	if err != nil {
		log.Error.Fatalln(err)
	}
	if val == "InProgress" {
		return true
	}
	return false
}
