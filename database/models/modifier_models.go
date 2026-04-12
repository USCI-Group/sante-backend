package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type InputType string

const (
	InputTypeRadio    InputType = "radio"
	InputTypeCheckbox InputType = "checkbox"
)

type DependencyType string

const (
	DependencyTypeDependent   DependencyType = "dependent"
	DependencyTypeIndependent DependencyType = "independent"
)

type ProductModifierMapping struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	ProductID       uuid.UUID       `json:"product_id" gorm:"type:uuid; index; nullable"`
	Product         *Product        `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	ModifierGroupID uuid.UUID       `json:"modifier_group_id" gorm:"type:uuid; index; nullable"`
	ModifierGroup   *ModifierGroups `json:"modifier_group,omitempty" gorm:"foreignKey:ModifierGroupID"`
	MaxSelection    int             `json:"max_selection" gorm:"type:integer; default:1"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       *time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `json:"deleted_at,omitempty"`
}

type ModifierGroups struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID     uuid.UUID       `json:"business_id" gorm:"type:uuid; index; nullable"`
	Business       *Business       `json:"business,omitempty" gorm:"foreignKey:BusinessID"`
	Name           string          `json:"name" gorm:"type:varchar(255)"`
	InputType      InputType       `json:"input_type" gorm:"type:varchar(255)"`
	DependencyType *DependencyType `json:"dependency_type" gorm:"type:varchar(50)"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      *time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `json:"deleted_at,omitempty"`
}

type ModifierOptions struct {
	ID                 uuid.UUID                   `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	ModifierGroupID    uuid.UUID                   `json:"modifier_group_id" gorm:"type:uuid; index; nullable"`
	ModifierGroup      *ModifierGroups             `json:"modifier_group,omitempty" gorm:"foreignKey:ModifierGroupID"`
	Name               string                      `json:"name" gorm:"type:varchar(255)"`
	PriceAdjustment    float32                     `json:"price_adjustment" gorm:"type:numeric"`
	IngredientMappings []ModifierIngredientMapping `json:"ingredient_mappings"`
	SortOrder          int                         `json:"sort_order" gorm:"type:integer; default:0"`
	IsActive           bool                        `json:"is_active" gorm:"type:boolean; default:true"`
	ImageURL           *string                     `json:"image_url" gorm:"type:varchar(255)"`
	CreatedAt          time.Time                   `json:"created_at"`
	UpdatedAt          *time.Time                  `json:"updated_at"`
	DeletedAt          gorm.DeletedAt              `json:"deleted_at,omitempty"`
}
