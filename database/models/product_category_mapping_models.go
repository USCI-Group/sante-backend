package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type ProductCategoryMapping struct {
	ID                   uuid.UUID          `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	ProductID            uuid.UUID          `json:"product_id" gorm:"type:uuid; index;nullable"`
	Product              Product            `json:"product" gorm:"foreignKey:ProductID"`
	ProductCategoryID    uuid.UUID          `json:"product_category_id" gorm:"type:uuid; index;nullable"`
	ProductCategory      ProductCategory    `json:"product_category" gorm:"foreignKey:ProductCategoryID"`
	ProductSubCategoryID *uuid.UUID         `json:"product_sub_category_id" gorm:"type:uuid; index;nullable"`
	ProductSubCategory   ProductSubCategory `json:"product_sub_category" gorm:"foreignKey:ProductSubCategoryID"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            *time.Time         `json:"updated_at"`
	DeletedAt            gorm.DeletedAt     `json:"deleted_at,omitempty"`
}
