package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type PointRedemptionTransaction struct {
	ID                    uuid.UUID           `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	PointRedemptionRuleID uuid.UUID           `json:"point_redemption_rule_id" gorm:"type:uuid"`
	PointRedemptionRule   PointRedemptionRule `json:"point_redemption_rule" gorm:"foreignKey:PointRedemptionRuleID"`
	CustomerID            uuid.UUID           `json:"customer_id" gorm:"type:uuid"`
	Customer              Customer            `json:"customer" gorm:"foreignKey:CustomerID"`
	PointsRedeemed        int                 `json:"points_redeemed"`
	RedeemedAt            time.Time           `json:"redeemed_at" gorm:"type:timestamp with time zone"`
	CreatedAt             time.Time           `json:"created_at"`
	UpdatedAt             *time.Time          `json:"updated_at"`
	DeletedAt             gorm.DeletedAt      `json:"deleted_at,omitempty"`
}
