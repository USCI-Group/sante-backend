package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type DiscountType string

const (
	DiscountTypePercent DiscountType = "percentage"
	DiscountTypeFixed   DiscountType = "fixed"
)

// Discount represents a discount or promotion.
type Discount struct {
	ID           uuid.UUID   `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	MembershipID *uuid.UUID  `json:"membership_id" gorm:"type:uuid"`
	Membership   *Membership `json:"membership" gorm:"foreignKey:MembershipID"`
	//BusinessID       *uuid.UUID     `json:"business_id" gorm:"type:uuid"`
	Name             string         `json:"name" gorm:"type:varchar(255)"`
	Description      string         `json:"description" gorm:"type:text"`
	DiscountType     DiscountType   `json:"discount_type" gorm:"type:varchar(50)"`
	Value            float32        `json:"value" gorm:"type:decimal(10,2)"`
	IsStackable      bool           `json:"is_stackable" gorm:"type:boolean"`
	UsageType        string         `json:"usage_type" gorm:"type:varchar(50)"`                  // e.g., "single", "multiple"
	MaxDiscountValue *float32       `json:"max_discount_value" gorm:"type:numeric;default:NULL"` // e.g., 100.00 for max $100 discount
	ValidFrom        time.Time      `json:"valid_from"`
	ValidTo          time.Time      `json:"valid_to"`
	IsActive         bool           `json:"is_active" gorm:"type:boolean"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        *time.Time     `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty"`
}
