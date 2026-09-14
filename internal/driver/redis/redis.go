package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/mcp-x/internal/driver"
)

type RedisDriver struct {
	client redis.UniversalClient
}

func (d *RedisDriver) Name() string            { return "redis" }
func (d *RedisDriver) Type() driver.DriverType { return driver.DriverTypeNoSQL }

func (d *RedisDriver) Connect(ctx context.Context, cfg driver.ConnConfig) error {
	switch cfg.Mode {
	case "cluster":
		d.client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    cfg.Addrs,
			Password: cfg.Password,
			PoolSize: cfg.PoolSize,
		})
	default:
		d.client = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
			PoolSize: cfg.PoolSize,
		})
	}
	if err := d.client.Ping(ctx).Err(); err != nil {
		d.client.Close()
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}

func (d *RedisDriver) Get(ctx context.Context, key string) (string, error) {
	val, err := d.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "(nil)", nil
	}
	return val, err
}

func (d *RedisDriver) Set(ctx context.Context, key string, value string, ttlSeconds int) error {
	if ttlSeconds > 0 {
		return d.client.Set(ctx, key, value, time.Duration(ttlSeconds)*time.Second).Err()
	}
	return d.client.Set(ctx, key, value, 0).Err()
}

func (d *RedisDriver) Del(ctx context.Context, key string) (int64, error) {
	return d.client.Del(ctx, key).Result()
}

func (d *RedisDriver) Keys(ctx context.Context, pattern string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	var keys []string
	iter := d.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= limit {
			break
		}
	}
	return keys, iter.Err()
}

func (d *RedisDriver) KeyType(ctx context.Context, key string) (string, error) {
	t, err := d.client.Type(ctx, key).Result()
	return string(t), err
}

func (d *RedisDriver) TTL(ctx context.Context, key string) (int64, error) {
	dur, err := d.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int64(dur.Seconds()), nil
}

func (d *RedisDriver) Execute(ctx context.Context, command string, args ...string) (any, error) {
	cmd := strings.ToUpper(strings.TrimSpace(command))
	intArgs := make([]any, 0, len(args)+1)
	intArgs = append(intArgs, cmd)
	for _, a := range args {
		intArgs = append(intArgs, a)
	}
	return d.client.Do(ctx, intArgs...).Result()
}

func (d *RedisDriver) Ping(ctx context.Context) error {
	return d.client.Ping(ctx).Err()
}

func (d *RedisDriver) Close() error {
	if d.client == nil {
		return nil
	}
	return d.client.Close()
}

func init() {
	driver.Register("redis", func() driver.AnyDriver { return &RedisDriver{} })
}
