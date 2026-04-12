package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type DeliveryType string

const (
	DeliveryTypeDelivery DeliveryType = "delivery"
	DeliveryTypePickup   DeliveryType = "pickup"
)

const (
	LeavingReasonFeedbackQuestion = "What is the reason for leaving?"
)

type Onboarding struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID  uuid.UUID  `json:"business_id" gorm:"type:uuid"`
	Business    *Business  `json:"business,omitempty" gorm:"foreignKey:BusinessID"`
	Title       string     `json:"title" gorm:"type:varchar(255)"`
	Description string     `json:"description,omitempty" gorm:"type:varchar(255)"`
	ImageURL    string     `json:"image_url,omitempty" gorm:"type:varchar(500)"`
	IsActive    bool       `json:"is_active" gorm:"type:boolean;"`
	SortOrder   int        `json:"sort_order" gorm:"type:int"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type Announcement struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID  uuid.UUID  `json:"business_id" gorm:"type:uuid"`
	Business    *Business  `json:"business,omitempty" gorm:"foreignKey:BusinessID"`
	Title       string     `json:"title" gorm:"type:varchar(255)"`
	Description string     `json:"description,omitempty" gorm:"type:varchar(500)"`
	ImageURL    string     `json:"image_url,omitempty" gorm:"type:varchar(500)"`
	IsActive    bool       `json:"is_active" gorm:"type:boolean;"`
	SortOrder   int        `json:"sort_order" gorm:"type:int"`
	StartDate   time.Time  `json:"start_date" gorm:"type:timestamp with time zone"`
	EndDate     time.Time  `json:"end_date" gorm:"type:timestamp with time zone"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type Delivery struct {
	ID           uuid.UUID    `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID   uuid.UUID    `json:"business_id" gorm:"type:uuid"`
	Business     *Business    `json:"business,omitempty" gorm:"foreignKey:BusinessID"`
	ImageURL     string       `json:"image_url,omitempty" gorm:"type:varchar(500)"`
	DeliveryType DeliveryType `json:"delivery_type" gorm:"type:varchar(255)"`
	IsActive     bool         `json:"is_active" gorm:"type:boolean;"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    *time.Time   `json:"updated_at"`
}

type FeedbackQuestion struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID uuid.UUID      `json:"business_id" gorm:"type:uuid"`
	Business   *Business      `json:"business,omitempty" gorm:"foreignKey:BusinessID"`
	Question   string         `json:"question" gorm:"type:varchar(500)"`
	Section    string         `json:"section" gorm:"type:varchar(255)"`
	ImageURL   string         `json:"image_url,omitempty" gorm:"type:varchar(500)"`
	IsActive   bool           `json:"is_active" gorm:"type:boolean;default:true"`
	SortOrder  int            `json:"sort_order" gorm:"type:int"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  *time.Time     `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type Feedback struct {
	ID                 uuid.UUID        `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	FeedbackQuestionID uuid.UUID        `json:"feedback_question_id" gorm:"type:uuid"`
	FeedbackQuestion   FeedbackQuestion `json:"feedback_question" gorm:"foreignKey:FeedbackQuestionID"`
	CustomerID         uuid.UUID        `json:"customer_id" gorm:"type:uuid"`
	Customer           Customer         `json:"customer" gorm:"foreignKey:CustomerID"`
	Rating             *int             `json:"rating" gorm:"type:int;check:rating >= 1 AND rating <= 5"`
	Comment            string           `json:"comment" gorm:"type:varchar(500)"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          *time.Time       `json:"updated_at"`
	DeletedAt          gorm.DeletedAt   `json:"deleted_at,omitempty"`
}
