package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type ProductCategory struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"type:varchar(255)" valid:"required~Name is required"`
	ImageURL    string         `json:"image_url" gorm:"type:varchar(500)"`
	BannerURL   string         `json:"banner_url" gorm:"type:varchar(500)"`
	Description string         `json:"description" gorm:"type:text"`
	BusinessID  *uuid.UUID     `json:"business_id" gorm:"type:uuid; index;nullable"`
	SortOrder   int            `json:"sort_order" gorm:"type:integer; default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type ProductSubCategory struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID  *uuid.UUID     `json:"business_id" gorm:"type:uuid; index;nullable"`
	Name        string         `json:"name" gorm:"type:varchar(255)" valid:"required~Name is required"`
	Description string         `json:"description" gorm:"type:text"`
	SortOrder   int            `json:"sort_order" gorm:"type:integer; default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}
