package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type CampaignPushNotification struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	Title      string         `json:"title" gorm:"type:varchar(255)"`
	Message    string         `json:"message" gorm:"type:varchar(255)"`
	BusinessID uuid.UUID      `json:"business_id" gorm:"type:uuid"`
	Business   *Business      `json:"business" gorm:"foreignKey:BusinessID"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty"`
}
