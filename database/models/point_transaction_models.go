package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type PointTransaction struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	PointRuleID  *uuid.UUID     `json:"point_rule_id" gorm:"type:uuid"`
	PointRule    *PointRule     `json:"point_rule" gorm:"foreignKey:PointRuleID"`
	CustomerID   uuid.UUID      `json:"customer_id" gorm:"type:uuid"`
	Customer     *Customer      `json:"customer" gorm:"foreignKey:CustomerID"`
	PointsEarned int            `json:"points_earned" gorm:"type:int"`
	OrderID      *uuid.UUID     `json:"order_id" gorm:"type:uuid" index:"order_id"`
	EarnedAt     time.Time      `json:"earned_at"`
	Details      *string        `json:"details" gorm:"type:varchar(255)"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    *time.Time     `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty"`
}
