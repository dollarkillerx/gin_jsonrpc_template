package client

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/google/feitian/pkg/common/config"
)

func RedisClient(conf config.RedisConfiguration) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.Db,
	})
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}
	return client, nil
}
