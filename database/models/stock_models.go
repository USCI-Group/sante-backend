package models

import (
	"time"

	"encore.app/common/constants"
	"encore.dev/types/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type StockRequestStatus string
type StockRequestResponseStatus string

const (
	StockRequestStatusNew      StockRequestStatus = "new"
	StockRequestStatusPending  StockRequestStatus = "pending"
	StockRequestStatusApproved StockRequestStatus = "approved"
	StockRequestStatusRejected StockRequestStatus = "rejected"
	// this status is used to cancel the request when the amendment is declined
	StockRequestStatusCancelled StockRequestStatus = "cancelled"
	StockRequestStatusCompleted StockRequestStatus = "completed"
)

const (
	StockRequestResponseStatusApproved StockRequestResponseStatus = "approved"
	StockRequestResponseStatusRejected StockRequestResponseStatus = "rejected"
	StockRequestResponseStatusAmended  StockRequestResponseStatus = "amended"
	// cancel mean decline the amendment request
	StockRequestResponseStatusCancelled StockRequestResponseStatus = "cancelled"
)

type Stock struct {
	ID                 uuid.UUID                 `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OutletID           uuid.UUID                 `json:"outlet_id" gorm:"type:uuid"`
	Outlet             *Outlet                   `json:"outlet" gorm:"foreignKey:OutletID"`
	IngredientID       uuid.UUID                 `json:"ingredient_id" gorm:"type:uuid"`
	Ingredient         *Ingredient               `json:"ingredient" gorm:"foreignKey:IngredientID"`
	Name               string                    `json:"name" gorm:"type:varchar(255)"`
	Description        string                    `json:"description" gorm:"type:text"`
	SmallScaleUnit     constants.UnitMeasurement `json:"small_scale_unit" gorm:"type:varchar(255)"`
	LargeScaleUnit     constants.UnitMeasurement `json:"large_scale_unit" gorm:"type:varchar(255)"`
	SmallScaleQuantity float32                   `json:"small_scale_quantity" gorm:"type:numeric"`
	LargeScaleQuantity float32                   `json:"large_scale_quantity" gorm:"type:numeric"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          *time.Time                `json:"updated_at"`
	DeletedAt          gorm.DeletedAt            `json:"deleted_at,omitempty"`
}

// stock request model (table)
type StockRequest struct {
	ID                uuid.UUID          `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	RequesterOutletID uuid.UUID          `json:"requester_outlet_id" gorm:"type:uuid"`
	RequesterOutlet   *Outlet            `json:"requester_outlet" gorm:"foreignKey:RequesterOutletID"`
	RequestDate       time.Time          `json:"request_date" gorm:"type:timestamp with time zone"`
	Remarks           string             `json:"remarks" gorm:"type:text"`
	RequestStatus     StockRequestStatus `json:"request_status" gorm:"type:varchar(255)"`
	RequesterID       uuid.UUID          `json:"requester_id" gorm:"type:uuid"`
	Requester         *User              `json:"requester" gorm:"foreignKey:RequesterID"`
	// this is the outlet that received the request
	ResponderOutletID *uuid.UUID                  `json:"responder_outlet_id" gorm:"type:uuid"`
	ResponderOutlet   *Outlet                     `json:"responder_outlet" gorm:"foreignKey:ResponderOutletID"`
	ResponderID       *uuid.UUID                  `json:"responder_id" gorm:"type:uuid"`
	Responder         *User                       `json:"responder" gorm:"foreignKey:ResponderID"`
	ResponseDate      *time.Time                  `json:"response_date" gorm:"type:timestamp with time zone"`
	ResponseStatus    *StockRequestResponseStatus `json:"response_status" gorm:"type:varchar(255)"`
	RequestedItems    []StockRequestedItem        `json:"requested_items" gorm:"foreignKey:StockRequestID"`
	ResponderRemarks  *string                     `json:"responder_remarks" gorm:"type:text"`
	IsReceived        bool                        `json:"is_received" gorm:"type:boolean;default:false"`
	CreatedAt         time.Time                   `json:"created_at"`
	UpdatedAt         *time.Time                  `json:"updated_at"`
	DeletedAt         gorm.DeletedAt              `json:"deleted_at,omitempty"`
}

// stock request item model (table)
type StockRequestedItem struct {
	ID             uuid.UUID                 `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	StockRequestID uuid.UUID                 `json:"stock_request_id" gorm:"type:uuid"`
	StockRequest   *StockRequest             `json:"stock_request" gorm:"foreignKey:StockRequestID"`
	IngredientID   uuid.UUID                 `json:"ingredient_id" gorm:"type:uuid"`
	Ingredient     *Ingredient               `json:"ingredient" gorm:"foreignKey:IngredientID"`
	UnitSelected   constants.UnitMeasurement `json:"unit_selected" gorm:"type:varchar(255)"`
	Quantity       decimal.Decimal           `json:"quantity" gorm:"type:decimal(20,6)"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      *time.Time                `json:"updated_at"`
	DeletedAt      gorm.DeletedAt            `json:"deleted_at,omitempty"`
}
