package redis_service

import (
	"context"
	"log"
	"time"

	"encore.app/common"
	"github.com/go-redis/redis"
)

//encore:service
type Service struct {
	redisClient *redis.Client
}

// It is automatically called by Encore on service startup.
func initService() (*Service, error) {
	address := common.GetRedisAddress()

	rdb := redis.NewClient(&redis.Options{
		Addr: address, // Change to your Redis server address if needed
		DB:   0,       // Use the default DB (DB 0)
	})
	// Test connection to Redis
	_, err := rdb.Ping().Result()
	if err != nil {
		log.Println("**")
		log.Println("**")
		log.Println("  Redis is not running!")
		log.Println("  Please follow the steps in 'sante-backend/README_CONFIG.md' and run redis.")
		log.Println("**")
		log.Println("**")

		if common.IsLocal() {
			log.Println("Continue running in local mode")
			return nil, err
		}
	}

	service := Service{redisClient: rdb}

	return &service, nil
}

// encore:api private method=POST path=/api/redis/auth-token/set
func (s *Service) SetAuthToken(ctx context.Context, req SetRequest) error {
	if req.Expiry != nil {
		expiry := time.Until((*req.Expiry).Add(-5 * time.Minute))

		return s.redisClient.Set(req.Key, req.Value, expiry).Err()
	}
	return s.redisClient.Set(req.Key, req.Value, 0).Err()
}

// encore:api private method=GET path=/api/redis/hgetall/:key
func (s *Service) HGetAll(ctx context.Context, key string) (*HGetAllResponse, error) {
	fields, err := s.redisClient.HGetAll(key).Result()
	if err != nil {
		return nil, err
	}

	return &HGetAllResponse{Fields: fields}, nil
}

// encore:api private method=POST path=/api/redis/hmset
func (s *Service) HMSet(ctx context.Context, req HMSetRequest) error {
	fields := make(map[string]interface{})
	for key, value := range req.Fields {
		fields[key] = value
	}
	return s.redisClient.HMSet(req.Key, fields).Err()
}

// encore:api private method=POST path=/api/redis/set
func (s *Service) Set(ctx context.Context, req SetRequest) error {
	if req.Expiry != nil {
		expiry := time.Until(*req.Expiry)
		return s.redisClient.Set(req.Key, req.Value, expiry).Err()
	}
	return s.redisClient.Set(req.Key, req.Value, 0).Err()
}

// encore:api private method=GET path=/api/redis/get/:key
func (s *Service) Get(ctx context.Context, key string) (*GetResponse, error) {
	value, err := s.redisClient.Get(key).Result()
	if err != nil {
		return nil, err
	}
	return &GetResponse{Value: &value}, nil
}

// encore:api private method=DELETE path=/api/redis/delete/:key
func (s *Service) Delete(ctx context.Context, key string) error {
	return s.redisClient.Del(key).Err()
}

// encore:api private method=POST path=/api/redis/push-queue
func (s *Service) PushQueue(ctx context.Context, req PushQueueRequest) error {
	return s.redisClient.LPush(req.QueueName, req.Data).Err()
}

// encore:api private method=GET path=/api/redis/pop-queue/:queueName
func (s *Service) PopQueue(ctx context.Context, queueName string) (*GetResponse, error) {
	data, err := s.redisClient.RPop(queueName).Result()
	if err != nil {
		return nil, err
	}
	return &GetResponse{Value: &data}, nil

}
