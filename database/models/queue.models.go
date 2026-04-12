package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type QueueType string

const (
	QueueTypeDineIn      QueueType = "dine_in"
	QueueTypeTakeAway    QueueType = "take_away"
	QueueTypeDelivery    QueueType = "delivery"
	QueueTypePickup      QueueType = "pickup"
	QueueTypePickupLater QueueType = "pickup_later"
	QueueTypeOnlineOrder QueueType = "online_order"
)

type QueueStatus string

const (
	QueueStatusPending    QueueStatus = "pending"
	QueueStatusProcessing QueueStatus = "processing"
	QueueStatusCompleted  QueueStatus = "completed"
	QueueStatusFailed     QueueStatus = "failed"
	QueueStatusCancelled  QueueStatus = "cancelled"
)

type NotificationQueue struct {
	ID               uuid.UUID        `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	QueueType        QueueType        `json:"queue_type" gorm:"type:varchar(100);index:idx_queue_type"`
	QueueStatus      QueueStatus      `json:"queue_status" gorm:"type:varchar(100);index:idx_queue_status"`
	Title            string           `json:"title" gorm:"type:varchar(255)"`
	Body             string           `json:"body" gorm:"type:text"`
	ActionURL        string           `json:"action_url" gorm:"type:text"`
	ImageURL         *string          `json:"image_url" gorm:"type:text"`
	NotificationType NotificationType `json:"notification_type" gorm:"type:text;index:idx_notification_type"`
	OrderID          uuid.UUID        `json:"order_id" gorm:"type:uuid;index:idx_order_id"`
	PlanToSendAt     time.Time        `json:"plan_to_send_at" gorm:"type:timestamp with time zone;index:idx_plan_to_send_at"` // plan to send at this time
	SendAt           *time.Time       `json:"send_at" gorm:"type:timestamp with time zone;index:idx_send_at"`                 // actually sent at this time
	CompletedAt      *time.Time       `json:"completed_at" gorm:"type:timestamp with time zone"`                              // actually completed at this time
	FailedAt         *time.Time       `json:"failed_at" gorm:"type:timestamp with time zone"`                                 // actually failed at this time
	CancelledAt      *time.Time       `json:"cancelled_at" gorm:"type:timestamp with time zone"`                              // actually cancelled at this time
	CreatedAt        time.Time        `json:"created_at" gorm:"type:timestamp with time zone;index:idx_created_at"`           // when the queue is created
	UpdatedAt        *time.Time       `json:"updated_at" gorm:"type:timestamp with time zone"`                                // when the queue is updated
	DeletedAt        gorm.DeletedAt   `json:"deleted_at,omitempty" gorm:"index:idx_deleted_at"`                               // when the queue is deleted
}
