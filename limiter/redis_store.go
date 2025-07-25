package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	Client *redis.Client
	Ctx    context.Context
}

func NewRedisStore(addr, password string) *RedisStore {
	return &RedisStore{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0,
		}),
		Ctx: context.Background(),
	}
}

func (r *RedisStore) Allow(key string, limit int, blockSeconds int) (bool, error) {
	count, err := r.Client.Incr(r.Ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		// set limit of key to expire in 1 second so liit is x req/sec
		r.Client.Expire(r.Ctx, key, time.Second)
	}

	if int(count) > limit {
		blockKey := fmt.Sprintf("block:%s", key)
		blocked, _ := r.Client.Get(r.Ctx, blockKey).Result()
		if blocked == "" {
			r.Client.Set(r.Ctx, blockKey, "1", time.Duration(blockSeconds)*time.Second)
		}
		return false, nil
	}

	blockKey := fmt.Sprintf("block:%s", key)
	blocked, _ := r.Client.Get(r.Ctx, blockKey).Result()
	if blocked != "" {
		return false, nil
	}

	return true, nil
}
