package redis_service

import "time"

type HGetAllResponse struct {
	Fields map[string]string `json:"fields"`
}

type HMSetRequest struct {
	Key    string            `json:"key"`
	Fields map[string]string `json:"fields"`
}

type SetRequest struct {
	Key    string     `json:"key"`
	Value  string     `json:"value"`
	Expiry *time.Time `json:"expiry"`
}

type GetResponse struct {
	Value *string `json:"value"`
}

type PushQueueRequest struct {
	QueueName string `json:"queue_name"`
	Data      []byte `json:"data"`
}
