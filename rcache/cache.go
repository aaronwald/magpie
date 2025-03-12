package rcache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

var (
	RedisUrl string = "redis:6379"
)

var redisContext = context.Background()

type EnableSound struct {
	Enabled bool `json:"enabled"`
}

func CheckEnabled() (EnableSound, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     RedisUrl,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	val, err := rdb.Get(redisContext, "sound_enabled").Result()
	if err != nil {
		slog.Error("check_enabled", "error", err)
		return EnableSound{Enabled: false}, err
	}

	var data EnableSound
	json.Unmarshal([]byte(val), &data)
	return data, nil
}

func SetOpenClose(topic string, status int) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     RedisUrl,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	val, err := rdb.Set(redisContext, topic, status, 0).Result()
	if err != nil {
		slog.Error("set_openclose", "error", err)
		return err
	}

	slog.Info("set_openclose", "topic", topic)
	slog.Info("set_openclose", "result", val)
	return nil
}

// func SetGarageState(topic string, state bool) error {
// 	rdb := redis.NewClient(&redis.Options{
// 		Addr:     RedisUrl,
// 		Password: "", // no password set
// 		DB:       0,  // use default DB
// 	})

// 	val, err := rdb.Set(redisContext, topic, state, 0).Result()
// 	if err != nil {
// 		slog.Error("set_garage_state", "error", err)
// 		return err
// 	}

// 	slog.Debug("set_garage_state", "topic", topic)
// 	slog.Debug("set_garage_state", "result", val)
// 	return nil
// }

func SubscribeRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     RedisUrl,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	soundCheck := EnableSound{Enabled: false}
	subscriber := rdb.Subscribe(redisContext, "magpie")
	for {
		msg, err := subscriber.ReceiveMessage(redisContext)
		if err != nil {
			panic(err)
		}

		if err := json.Unmarshal([]byte(msg.Payload), &soundCheck); err != nil {
			panic(err)
		}

		fmt.Println("Received message from " + msg.Channel + " channel.")
		fmt.Printf("%+v\n", soundCheck)
	}
}
