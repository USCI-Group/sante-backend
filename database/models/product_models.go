package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID                   uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID           *uuid.UUID     `json:"business_id" gorm:"type:uuid; index;nullable"`
	Name                 string         `json:"name" gorm:"type:varchar(255)" valid:"required~Name is required"`
	Description          string         `json:"description" gorm:"type:text" `
	Cost                 float32        `json:"cost" gorm:"type:numeric" validate:"gte=0"`
	BasePrice            float32        `json:"base_price" gorm:"type:numeric" validate:"gte=0"`
	Price                float32        `json:"price" gorm:"type:numeric" validate:"gte=0"`
	IsStandalonePurchase bool           `json:"is_standalone_purchase" gorm:"type:boolean; default:true"`
	IsAddon              bool           `json:"is_addon" gorm:"type:boolean; default:false"`
	ImageURL             string         `json:"image_url" gorm:"type:varchar(255)"`
	IsActive             bool           `json:"is_active" gorm:"type:boolean; default:true"`
	IsStoreOutlet        bool           `json:"is_store_outlet" gorm:"type:boolean; default:true"`
	IsGrabFood           bool           `json:"is_grab_food" gorm:"type:boolean; default:false"`
	IsShopeeFood         bool           `json:"is_shopee_food" gorm:"type:boolean; default:false"`
	GrabFoodInfo         GrabFoodInfo   `json:"grab_food_info" gorm:"embedded"`
	ShopeeFoodInfo       ShopeeFoodInfo `json:"shopee_food_info" gorm:"embedded"`
	SortOrder            int            `json:"sort_order" gorm:"type:integer; default:0"`
	ModifierOptionsID    *uuid.UUID     `json:"modifier_options_id" gorm:"type:uuid;index;nullable"`
	ExperiencePoints     int            `json:"experience_points" gorm:"type:integer; default:0"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            *time.Time     `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type GrabFoodInfo struct {
	GrabFoodPrice *float32 `json:"grab_food_price" gorm:"type:numeric"`
}
type ShopeeFoodInfo struct {
	ShopeeFoodPrice *float32 `json:"shopee_food_price" gorm:"type:numeric"`
}
