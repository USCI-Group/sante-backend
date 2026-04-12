package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	ID                  uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OrderID             uuid.UUID      `json:"order_id" gorm:"type:uuid"`
	Order               Order          `json:"order" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	TransactionNumber   *string        `json:"transaction_number" gorm:"type:varchar(255);uniqueIndex"`
	MolTransactionID    *string        `json:"mol_transaction_id" gorm:"type:varchar(255);uniqueIndex"` // Deprecated for new providers
	GatewayTransactionID *string        `json:"gateway_transaction_id" gorm:"type:varchar(255);uniqueIndex"`
	PaymentURL          *string        `json:"payment_url" gorm:"type:text"`
	MerchantReference   *string        `json:"merchant_reference" gorm:"type:varchar(255)"`
	TransactionDate     time.Time      `json:"transaction_date" gorm:"type:timestamp with time zone"`
	Amount              float32        `json:"amount" gorm:"type:decimal(10,2)"`
	PaymentMethod       string         `json:"payment_method" gorm:"type:varchar(255)"`
	PaymentStatus       string         `json:"payment_status" gorm:"type:varchar(255)"`
	ErrorCode           string         `json:"error_code" gorm:"type:varchar(255)"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           *time.Time     `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"deleted_at,omitempty"`
}

