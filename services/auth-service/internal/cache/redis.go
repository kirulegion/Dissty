package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// OTPCache handles storing and retrieving OTP codes in Redis
type OTPCache struct {
	client *redis.Client
}

func NewOTPCache(client *redis.Client) *OTPCache {
	return &OTPCache{client: client}
}

// Set stores an OTP code for an identifier with 10 minute expiry
func (c *OTPCache) Set(ctx context.Context, identifier, code string) error {
	key := fmt.Sprintf("otp:%s", identifier)
	return c.client.Set(ctx, key, code, 10*time.Minute).Err()
}

// Get retrieves an OTP code for an identifier
func (c *OTPCache) Get(ctx context.Context, identifier string) (string, error) {
	key := fmt.Sprintf("otp:%s", identifier)
	return c.client.Get(ctx, key).Result()
}

// Delete removes an OTP code after successful verification
func (c *OTPCache) Delete(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("otp:%s", identifier)
	return c.client.Del(ctx, key).Err()
}



