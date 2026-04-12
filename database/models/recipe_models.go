package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type Recipe struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	ProductID   uuid.UUID      `json:"product_id" gorm:"type:uuid; index; nullable"`
	Product     Product        `json:"product" gorm:"foreignKey:ProductID"`
	Name        string         `json:"name" gorm:"type:varchar(255)" valid:"required~Name is required"`
	Description string         `json:"description" gorm:"type:text" `
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type RecipeStep struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	RecipeID    uuid.UUID      `json:"recipe_id" gorm:"type:uuid; index; nullable"`
	Recipe      Recipe         `json:"recipe" gorm:"foreignKey:RecipeID"`
	Name        string         `json:"name" gorm:"type:varchar(255)" valid:"required~Name is required"`
	Instruction string         `json:"instruction" gorm:"type:text" `
	Precedence  int            `json:"precedence" gorm:"type:integer" `
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}
