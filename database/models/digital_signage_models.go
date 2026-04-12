package models

import (
	"time"

	"encore.app/common/constants"
	"encore.dev/types/uuid"
)

type DigitalSignageContent struct {
	ID          uuid.UUID             `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID  uuid.UUID             `json:"business_id" gorm:"type:uuid"`
	ContentType constants.ContentType `json:"content_type" gorm:"type:varchar(255)" valid:"required~Content type is required"`
	ContentURL  string                `json:"content_url" gorm:"type:varchar(500)"`
	SortOrder   int                   `json:"sort_order" gorm:"type:int; default:0"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   *time.Time            `json:"updated_at"`
}

type DigitalSignageSlide struct {
	ID                       uuid.UUID                 `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID               uuid.UUID                 `json:"business_id" gorm:"type:uuid"`
	Title                    string                    `json:"title" gorm:"type:varchar(255)" valid:"required~Title is required,maxstringlength(30)~Title must not exceed 30 characters"`
	DigitalSignageSlideItems []DigitalSignageSlideItem `json:"digital_signage_slide_items" gorm:"foreignKey:DigitalSignageSlideID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt                time.Time                 `json:"created_at"`
	UpdatedAt                *time.Time                `json:"updated_at"`
}

type DigitalSignageSlideItem struct {
	ID                    uuid.UUID             `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	ContentID             uuid.UUID             `json:"content_id" gorm:"type:uuid"`
	Content               DigitalSignageContent `json:"content" gorm:"foreignKey:ContentID;references:ID"`
	DigitalSignageSlideID uuid.UUID             `json:"digital_signage_slide_id" gorm:"type:uuid; not null"`
	SortOrder             int                   `json:"sort_order" gorm:"type:int; default:0"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             *time.Time            `json:"updated_at"`
}
