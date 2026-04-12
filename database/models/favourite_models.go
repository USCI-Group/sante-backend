package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type CustomerFavouriteProduct struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	CustomerID uuid.UUID      `json:"customer_id" gorm:"type:uuid"`
	ProductID  uuid.UUID      `json:"product_id" gorm:"type:uuid"`
	Product    Product        `json:"product" gorm:"foreignKey:ProductID"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  *time.Time     `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty"`
}
