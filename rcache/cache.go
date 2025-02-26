package rcache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

var (
	redisUri string = "redis:6379"
)

var redisContext = context.Background()

type EnableSound struct {
	Enabled bool `json:"enabled"`
}

func CheckEnabled() (EnableSound, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisUri,
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

func SubscribeRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisUri,
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
