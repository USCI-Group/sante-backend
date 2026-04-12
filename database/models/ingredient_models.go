package models

import (
	"time"

	"encore.app/common/constants"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// Ingredient is a model that represents an ingredient item for a business
type Ingredient struct {
	ID           uuid.UUID                 `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID   uuid.UUID                 `json:"business_id" gorm:"type:uuid"`
	Business     *Business                 `json:"business" gorm:"foreignKey:BusinessID"`
	Name         string                    `json:"name" gorm:"type:varchar(255)"`
	Description  string                    `json:"description" gorm:"type:text"`
	Unit         constants.UnitMeasurement `json:"unit" gorm:"type:varchar(255)"`
	Quantity     float32                   `json:"quantity" gorm:"type:numeric"`
	PricePerUnit float32                   `json:"price_per_unit" gorm:"type:numeric"`
	ImageURL     string                    `json:"image_url" gorm:"type:varchar(255)"`
	SortOrder    int                       `json:"sort_order" gorm:"type:integer; default:0"`
	CreatedAt    time.Time                 `json:"created_at"`
	UpdatedAt    *time.Time                `json:"updated_at"`
	DeletedAt    gorm.DeletedAt            `json:"deleted_at,omitempty"`
}

// link product with ingredients (one product can have multiple ingredients)
type ProductIngredientMapping struct {
	ID           uuid.UUID                 `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	ProductID    uuid.UUID                 `json:"product_id" gorm:"type:uuid; index; nullable"`
	Product      Product                   `json:"product" gorm:"foreignKey:ProductID"`
	IngredientID uuid.UUID                 `json:"ingredient_id" gorm:"type:uuid; index; nullable"`
	Ingredient   Ingredient                `json:"ingredient" gorm:"foreignKey:IngredientID"`
	Unit         constants.UnitMeasurement `json:"unit" gorm:"type:varchar(255)"`
	Quantity     float32                   `json:"quantity" gorm:"type:numeric"`
	CreatedAt    time.Time                 `json:"created_at"`
	UpdatedAt    *time.Time                `json:"updated_at"`
	DeletedAt    gorm.DeletedAt            `json:"deleted_at,omitempty"`
}

type ModifierIngredientMapping struct {
	ID                uuid.UUID                 `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	IngredientID      uuid.UUID                 `json:"ingredient_id" gorm:"type:uuid; index; nullable"`
	Ingredient        Ingredient                `json:"ingredient" gorm:"foreignKey:IngredientID"`
	ModifierOptionsID uuid.UUID                 `json:"modifier_options_id" gorm:"type:uuid; index;"`
	Unit              constants.UnitMeasurement `json:"unit" gorm:"type:varchar(255)"`
	Quantity          float32                   `json:"quantity" gorm:"type:numeric"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         *time.Time                `json:"updated_at"`
	DeletedAt         gorm.DeletedAt            `json:"deleted_at,omitempty"`
}
