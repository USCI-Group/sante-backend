package models

import (
	"time"

	"encore.dev/types/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductWastageType struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID  uuid.UUID      `json:"business_id" gorm:"type:uuid"`
	Name        string         `json:"name" gorm:"type:varchar(255)" validate:"required"`
	Description string         `json:"description" gorm:"type:text"`
	IsActive    bool           `json:"is_active" gorm:"type:boolean; default:true"`
	SortOrder   int            `json:"sort_order" gorm:"type:integer; default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type StockReport struct {
	ID              uuid.UUID        `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID      uuid.UUID        `json:"business_id" gorm:"type:uuid"`
	OutletID        uuid.UUID        `json:"outlet_id" gorm:"type:uuid"`
	IngredientID    uuid.UUID        `json:"ingredient_id" gorm:"type:uuid"`
	Ingredient      *Ingredient      `json:"ingredient" gorm:"foreignKey:IngredientID"`
	Sales           float32          `json:"sales" gorm:"type:numeric"`
	Purchases       float32          `json:"purchases" gorm:"type:numeric"`
	TransferIn      *decimal.Decimal `json:"transfer_in" gorm:"type:numeric"`
	TransferOut     *decimal.Decimal `json:"transfer_out" gorm:"type:numeric"`
	Wastage         float32          `json:"wastage" gorm:"type:numeric"`
	Opening         *decimal.Decimal `json:"opening" gorm:"type:decimal(20,6)"`
	OpeningBySystem *decimal.Decimal `json:"opening_by_system" gorm:"type:decimal(20,6)"`
	Closing         *decimal.Decimal `json:"closing" gorm:"type:decimal(20,6)"`
	ClosingBySystem *decimal.Decimal `json:"closing_by_system" gorm:"type:decimal(20,6)"`
	Variance        *decimal.Decimal `json:"variance" gorm:"type:decimal(20,6)"`
	CashOpening     *decimal.Decimal `json:"cash_opening" gorm:"type:decimal(20,2)"`
	CashClosing     *decimal.Decimal `json:"cash_closing" gorm:"type:decimal(20,2)"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       *time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `json:"deleted_at,omitempty"`
}

type ProductWastageReport struct {
	ID            uuid.UUID           `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID    uuid.UUID           `json:"business_id" gorm:"type:uuid"`
	OutletID      uuid.UUID           `json:"outlet_id" gorm:"type:uuid"`
	ProductID     uuid.UUID           `json:"product_id" gorm:"type:uuid"`
	Product       *Product            `json:"product" gorm:"foreignKey:ProductID"`
	WastageTypeID *uuid.UUID          `json:"wastage_type_id" gorm:"type:uuid; nullable"`
	WastageType   *ProductWastageType `json:"wastage_type" gorm:"foreignKey:WastageTypeID"`
	WastageAmount float32             `json:"wastage_amount" gorm:"type:decimal(20,6)"`
	ReportDate    time.Time           `json:"report_date" gorm:"type:timestamp with time zone"`
	Notes         string              `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     *time.Time          `json:"updated_at"`
	DeletedAt     gorm.DeletedAt      `json:"deleted_at,omitempty"`
}
