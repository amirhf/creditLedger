package idem

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type Guard struct{ rdb *redis.Client }

func NewGuard(url string) *Guard {
	opt, _ := redis.ParseURL(url)
	return &Guard{rdb: redis.NewClient(opt)}
}

func (g *Guard) Claim(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	res, err := g.rdb.SetNX(ctx, key, "PENDING", ttl).Result()
	return res, err
}
