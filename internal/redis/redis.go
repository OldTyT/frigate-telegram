package redis

import (
	"context"
	"time"

	"github.com/oldtyt/frigate-telegram/internal/config"
	"github.com/oldtyt/frigate-telegram/internal/log"
	redis "github.com/redis/go-redis/v9"
)

var (
	RedisKeyStateSendEvent string          = "FrigateTelegramStoptSendEventMessage"
	RedisKeyStateMuteEvent string          = "FrigateTelegramStoptSendEventMessage"
	ctx                    context.Context = context.Background()
	conf                   *config.Config  = config.New()
)

var rdb = redis.NewClient(&redis.Options{
	Addr:     conf.RedisAddr,
	Password: conf.RedisPassword, // no password set
	DB:       conf.RedisDB,       // use default DB
	Protocol: conf.RedisProtocol, // specify 2 for RESP 2 or 3 for RESP 3
})

// Set state send event msg in redis
func SetStateSendEvent(send bool) bool {
	// send = true - send msg
	// send = false - don't send msg
	if send {
		err := rdb.Set(ctx, RedisKeyStateSendEvent, 1, 0).Err()
		if err != nil {
			log.Error.Fatalln(err)
			return false
		}
		return true
	} else {
		rdb.Del(ctx, RedisKeyStateSendEvent)
		return true
	}
}

// Get state send event msg from redis
func GetStateSendEvent() bool {
	// bool = false - send msg
	// bool = true - don't send msg
	_, err := rdb.Get(ctx, RedisKeyStateSendEvent).Result()
	//nolint:gosimple
	if err != nil {
		return true
	}

	return false
}

// Set state notify event msg in redis
func SetStateMuteEvent(mute bool) bool {
	// mute = true - send mute event
	// mute = false - don't mute event msg
	if mute {
		err := rdb.Set(ctx, RedisKeyStateMuteEvent, 1, 0).Err()
		if err != nil {
			log.Error.Fatalln(err)
			return false
		}
		return true
	} else {
		rdb.Del(ctx, RedisKeyStateMuteEvent)
		return true
	}
}

// Get state send event msg from redis
func GetStateMuteEvent() bool {
	// mute = true - send mute event
	// mute = false - don't mute event msg
	_, err := rdb.Get(ctx, RedisKeyStateMuteEvent).Result()
	//nolint:gosimple
	if err != nil {
		return false
	}
	return true
}

func AddNewEvent(EventID string, State string, RedisTTL time.Duration) {
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
	if val == "Finished" {
		return false
	}
	if val == "InWork" {
		return false
	}
	return false
}
