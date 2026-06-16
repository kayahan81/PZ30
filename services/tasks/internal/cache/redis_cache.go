package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client    *redis.Client
	baseTTL   time.Duration
	jitterSec int
	available bool
	mu        sync.RWMutex
}

type CacheConfig struct {
	Addr           string
	Password       string
	BaseTTLSeconds int
	JitterSeconds  int
}

func NewRedisCache(cfg CacheConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           0,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		DialTimeout:  5 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	rc := &RedisCache{
		client:    client,
		baseTTL:   time.Duration(cfg.BaseTTLSeconds) * time.Second,
		jitterSec: cfg.JitterSeconds,
		available: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("[WARN] Redis not available: %v", err)
		rc.available = false
		return rc, nil
	}

	log.Printf("[INFO] Redis connected to %s", cfg.Addr)
	return rc, nil
}

func (r *RedisCache) GetTaskKey(taskID string) string {
	return fmt.Sprintf("tasks:task:%s", taskID)
}

func (r *RedisCache) GetListKey() string {
	return "tasks:list"
}

func (r *RedisCache) calculateTTL() time.Duration {
	jitter := time.Duration(rand.Intn(r.jitterSec)) * time.Second
	return r.baseTTL + jitter
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	r.mu.RLock()
	if !r.available || r.client == nil {
		r.mu.RUnlock()
		return nil, fmt.Errorf("redis not available")
	}
	r.mu.RUnlock()

	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		log.Printf("[WARN] Redis GET error for key %s: %v", key, err)
		return nil, err
	}
	log.Printf("[DEBUG] Redis GET hit for key %s", key)
	return val, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	r.mu.RLock()
	if !r.available || r.client == nil {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal value for key %s: %v", key, err)
		return err
	}

	actualTTL := ttl
	if actualTTL == 0 {
		actualTTL = r.calculateTTL()
	}

	log.Printf("[DEBUG] Redis SET key %s with TTL %v", key, actualTTL)
	return r.client.Set(ctx, key, data, actualTTL).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	r.mu.RLock()
	if !r.available || r.client == nil {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	log.Printf("[DEBUG] Redis DEL key %s", key)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

func (r *RedisCache) IsAvailable() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.available || r.client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		r.mu.Lock()
		r.available = false
		r.mu.Unlock()
		return false
	}

	return true
}

func (r *RedisCache) BaseTTL() time.Duration {
	return r.baseTTL
}
