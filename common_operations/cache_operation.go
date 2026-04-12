package common_operations

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
)

// set cache in redis
func SetCache(redisClient *redis.Client, key string, value interface{}, expiration time.Duration) error {
	// Marshal the value to JSON before storing
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return redisClient.Set(key, jsonData, expiration).Err()
}

// get cache from redis
func GetCache[T any](redisClient *redis.Client, key string) (T, error) {
	var value T
	// Get the raw JSON string from Redis
	jsonData, err := redisClient.Get(key).Result()
	if err != nil {
		return value, err
	}
	// Unmarshal JSON to the target type
	err = json.Unmarshal([]byte(jsonData), &value)
	if err != nil {
		return value, err
	}
	return value, nil
}

// delete cache from redis
func DeleteCache(redisClient *redis.Client, key string) error {
	return redisClient.Del(key).Err()
}
