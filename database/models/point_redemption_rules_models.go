package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// one businesss can only have one point redemption rule
// applicable tier level based on number precedence
// RedemptionRule represents the rules for redeeming points
type PointRedemptionRule struct {
	ID                          uuid.UUID `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID                  uuid.UUID `json:"business_id" gorm:"type:uuid"`
	Business                    Business  `json:"business" gorm:"foreignKey:BusinessID"`
	RuleName                    string    `json:"rule_name" gorm:"type:varchar(255)"`
	Description                 string    `json:"description" gorm:"type:varchar(255)"`
	TermsAndConditions          string    `json:"terms_and_conditions" gorm:"type:text"`
	MinAmount                   float32   `json:"min_amount" gorm:"type:numeric" validate:"gte=0"`
	MaxAmount                   float32   `json:"max_amount" gorm:"type:numeric" validate:"gte=0"`
	MaxRedemption               int       `json:"max_redemption"`
	MaxRedemptionPerCustomer    int       `json:"max_redemption_per_customer"`
	MaxRedemptionPerDay         int       `json:"max_redemption_per_day"`
	MaxRedemptionPerMonth       int       `json:"max_redemption_per_month"`
	MaxRedemptionPerYear        int       `json:"max_redemption_per_year"`
	MaxRedemptionPerTransaction int       `json:"max_redemption_per_transaction"`
	MaxRedemptionPerOrder       int       `json:"max_redemption_per_order"`
	ApplicableTierLevel         int       `json:"applicable_tier_level"`
	IsActive                    bool      `json:"is_active"`
	ValidFrom                   time.Time `json:"valid_from"`
	ValidTo                     time.Time `json:"valid_to"`
	IsAllowAllPaymentType       bool      `json:"is_allow_all_payment_type"`
	//PaymentTypeAllowed          []string       `json:"payment_type_allowed"`
	//ApplicableTransactionType   []string       `json:"applicable_transaction_type"`
	DiscountPercentage float32        `json:"discount_percentage" gorm:"type:numeric" validate:"gte=0,lte=100"`
	ExchangeRate       float32        `json:"exchange_rate" gorm:"type:numeric" validate:"gte=0"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          *time.Time     `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at,omitempty"`
}
