package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store 验证码存储接口
type Store interface {
	Set(ctx context.Context, captchaID string, data interface{}, expireTime time.Duration) error
	Get(ctx context.Context, captchaID string) (string, error)
	Del(ctx context.Context, captchaID string) error
}

// RedisStore Redis 存储
type RedisStore struct {
	client *redis.Client
	prefix string
}

// NewRedisStore 创建 Redis 存储
func NewRedisStore(client *redis.Client, prefix string) *RedisStore {
	if prefix == "" {
		prefix = "captcha:"
	}
	return &RedisStore{
		client: client,
		prefix: prefix,
	}
}

// Set 存储验证码
func (s *RedisStore) Set(ctx context.Context, captchaID string, data interface{}, expireTime time.Duration) error {
	key := s.prefix + captchaID

	value, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal captcha data: %w", err)
	}

	err = s.client.Set(ctx, key, value, expireTime).Err()
	if err != nil {
		return fmt.Errorf("failed to set captcha: %w", err)
	}

	return nil
}

// Get 获取验证码
func (s *RedisStore) Get(ctx context.Context, captchaID string) (string, error) {
	key := s.prefix + captchaID

	value, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get captcha: %w", err)
	}

	return value, nil
}

// Del 删除验证码
func (s *RedisStore) Del(ctx context.Context, captchaID string) error {
	key := s.prefix + captchaID

	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete captcha: %w", err)
	}

	return nil
}

// MemStore 内存存储（用于测试）
type MemStore struct {
	data map[string]string
}

// NewMemStore 创建内存存储
func NewMemStore() *MemStore {
	return &MemStore{
		data: make(map[string]string),
	}
}

// Set 存储验证码
func (s *MemStore) Set(ctx context.Context, captchaID string, data interface{}, expireTime time.Duration) error {
	value, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal captcha data: %w", err)
	}

	s.data[captchaID] = string(value)

	// 自动过期
	go func() {
		time.Sleep(expireTime)
		delete(s.data, captchaID)
	}()

	return nil
}

// Get 获取验证码
func (s *MemStore) Get(ctx context.Context, captchaID string) (string, error) {
	value, ok := s.data[captchaID]
	if !ok {
		return "", ErrCaptchaNotFound
	}
	return value, nil
}

// Del 删除验证码
func (s *MemStore) Del(ctx context.Context, captchaID string) error {
	delete(s.data, captchaID)
	return nil
}
