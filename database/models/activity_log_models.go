package models

import (
	"time"

	"encore.app/common/constants"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type ActivityLog struct {
	ID             uuid.UUID                   `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	Activity       constants.LogAction         `json:"activity" gorm:"type:varchar(255)"`
	Status         constants.ActivityLogStatus `json:"status" gorm:"type:varchar(255)"`
	ActionByUserID uuid.UUID                   `json:"action_by_user_id" gorm:"type:uuid;default:null"`
	User           *User                       `json:"user,omitempty" gorm:"foreignKey:ActionByUserID"`
	ActionBy       string                      `json:"action_by" gorm:"type:varchar(255)"`
	Details        string                      `json:"details,omitempty" gorm:"type:text"`
	ErrorMessage   string                      `json:"error_message,omitempty" gorm:"type:varchar(255)"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      *time.Time                  `json:"updated_at"`
	DeletedAt      gorm.DeletedAt              `json:"deleted_at,omitempty"`
}

type Cache struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	Key       string         `json:"key" gorm:"type:varchar(255)"`
	Value     string         `json:"value" gorm:"type:text"`
	Expiry    *time.Time     `json:"expiry,omitempty" gorm:"type:timestamp with time zone"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}
