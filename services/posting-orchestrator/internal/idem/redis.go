package idem

import (
	"context"
	"log"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type Guard struct{ rdb *redis.Client }

func NewGuard(url string) *Guard {
	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Printf("Failed to parse Redis URL '%s', using default options: %v", url, err)
		// Fallback to default localhost options
		opt = &redis.Options{
			Addr: "localhost:6379",
			DB:   0,
		}
	}
	return &Guard{rdb: redis.NewClient(opt)}
}

func (g *Guard) Claim(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	res, err := g.rdb.SetNX(ctx, key, "PENDING", ttl).Result()
	return res, err
}
