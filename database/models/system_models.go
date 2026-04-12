package models

import (
	"time"

	"encore.app/common/constants"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// SystemData is a model that contains data for the system(sante)
type SystemData struct {
	ID          uuid.UUID                    `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	InfoType    constants.SystemDataInfoType `json:"info_type" gorm:"type:varchar(255);uniqueIndex:idx_info_type"`
	InfoValue   string                       `json:"info_value" gorm:"type:varchar(500)"`
	Expiry      *time.Time                   `json:"expiry" gorm:"type:timestamp with time zone"`
	IsEncrypted bool                         `json:"is_encrypted" gorm:"type:boolean;default:false"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   *time.Time                   `json:"updated_at"`
	DeletedAt   gorm.DeletedAt               `json:"deleted_at,omitempty"`
}
