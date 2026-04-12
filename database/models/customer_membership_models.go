package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type CustomerMembership struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	CustomerID   uuid.UUID      `json:"customer_id" gorm:"type:uuid"`
	Customer     *Customer      `json:"customer" gorm:"foreignKey:CustomerID"`
	MembershipID uuid.UUID      `json:"membership_id" gorm:"type:uuid"`
	Membership   *Membership    `json:"membership" gorm:"foreignKey:MembershipID"`
	ExpiryDate   *time.Time     `json:"expiry_date" gorm:"type:date"`
	Points       int            `json:"points" gorm:"type:int"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    *time.Time     `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty"`
}
