// Package cache membungkus Redis untuk seluruh aplikasi dengan namespace kunci yang konsisten.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient(addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("cache: failed to connect to redis: %w", err)
	}
	return client, nil
}

// Service adalah pembungkus bertipe agar domain tidak menyebar panggilan client.Get/Set mentah. Semua kunci lewat Key().
type Service struct {
	client *redis.Client
}

func NewService(client *redis.Client) *Service {
	return &Service{client: client}
}

// Key membuat kunci ber-namespace: kpr:<domain>:<id>. Contoh: Key(package, list:active) menghasilkan kpr:package:list:active.
func Key(domain string, parts ...string) string {
	key := "kpr:" + domain
	for _, p := range parts {
		key += ":" + p
	}
	return key
}

func (s *Service) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	raw, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, raw, ttl).Err()
}

func (s *Service) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}
