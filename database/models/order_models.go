package models

import (
	"time"

	"encore.app/common/constants"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type VoucherDiscountType string

const (
	OrderDiscountTypeFixed      VoucherDiscountType = "fixed"
	OrderDiscountTypePercentage VoucherDiscountType = "percentage"
	OrderDiscountTypeNone       VoucherDiscountType = "none"
)

// use for order status, order type, payment method, payment status
const (
	OrderStatusPending   = "pending"
	OrderStatusPreparing = "preparing"
	OrderStatusReady     = "ready"
	OrderStatusCollected = "collected"
	OrderStatusOnTheWay  = "on_the_way"
	OrderStatusDelivered = "delivered"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"

	OrderTypeDineIn      = "dine_in"
	OrderTypeTakeAway    = "take_away"
	OrderTypeDelivery    = "delivery"
	OrderTypePickup      = "pickup"
	OrderTypePickupLater = "pickup_later"
	OrderTypeOther       = "other"

	CasePaymentCash         = "cash"
	CasePaymentCreditCard   = "credit_card"
	CasePaymentDebitCard    = "debit_card"
	CasePaymentEWallet      = "e-wallet"
	CasePaymentStaticQR     = "static_qr"
	CasePaymentGrabCash     = "CASH"
	CasePaymentGrabCashless = "CASHLESS"
	CasePaymentOther        = "other"

	PaymentStatusPending          = "pending"
	PaymentStatusPendingAuthorize = "pending_authorize"
	PaymentStatusCompleted        = "completed" // when order is paid, then need to turn to completed if payment completed
	PaymentStatusFailed           = "failed"
	PaymentStatusRefunded         = "refunded" // when order is paid, then need to turn to refunded if refund completed
	PaymentStatusVoided           = "voided"   // when order is paid, then need to turn to voided if refund requested
)

var PaymentMethods = []string{
	CasePaymentCash,
	CasePaymentCreditCard,
	CasePaymentDebitCard,
	CasePaymentEWallet,
	CasePaymentStaticQR,
	CasePaymentGrabCash,
	CasePaymentGrabCashless,
}

type Platform string

const (
	PlatformGrabFood      Platform = "grabfood"
	PlatformShopeeFood    Platform = "shopeefood"
	PlatformStoreOutlet   Platform = "store_outlet"
	PlatformMembershipApp Platform = "membership_app"
)

type Order struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID    uuid.UUID  `json:"business_id" gorm:"type:uuid"`
	Business      *Business  `json:"business" gorm:"foreignKey:BusinessID"`
	OutletID      uuid.UUID  `json:"outlet_id" gorm:"type:uuid"`
	Outlet        *Outlet    `json:"outlet" gorm:"foreignKey:OutletID"`
	UserID        *uuid.UUID `json:"user_id" gorm:"type:uuid"`
	User          *User      `json:"user" gorm:"foreignKey:UserID"`
	CustomerID    *uuid.UUID `json:"customer_id" gorm:"type:uuid"`
	Customer      *Customer  `json:"customer" gorm:"foreignKey:CustomerID"`
	OrderNumber   string     `json:"order_number" gorm:"type:varchar(255)"`
	OrderDate     time.Time  `json:"order_date" gorm:"type:timestamp with time zone"`
	OrderType     string     `json:"order_type" gorm:"type:varchar(255)" atlas:"enum(dine_in,take_away,delivery)"`
	InvoiceNumber string     `json:"invoice_number" gorm:"type:varchar(255)"`
	InvoiceDate   time.Time  `json:"invoice_date" gorm:"type:timestamp with time zone"`
	// Order status (pending, payment, completed, cancelled)
	OrderStatus string `json:"order_status" gorm:"type:varchar(255)" enum:"pending,completed,cancelled" default:"pending"`
	// GrabFood Info
	Platform        *Platform `json:"platform" gorm:"type:varchar(255)" default:"store_outlet"`
	PlatformOrderID *string   `json:"platform_order_id" gorm:"type:varchar(255)"`
	PlatformState   *string   `json:"platform_state" gorm:"type:varchar(255)"`

	// TaxID              string         `json:"tax_id" gorm:"type:varchar(255)"`
	// TaxAmount          float32        `json:"tax_amount" gorm:"type:decimal(10,2)"`
	// gross total refer to the total amount of the order without tax, service charge, and discount
	GrossTotal                float32                   `json:"gross_total" gorm:"type:decimal(10,2)"`
	NetTotal                  float32                   `json:"net_total" gorm:"type:decimal(10,2)"`
	RoundedAmount             float32                   `json:"rounded_amount" gorm:"type:decimal(10,2)"`
	RoundedNetTotal           float32                   `json:"rounded_net_total" gorm:"type:decimal(10,2)"`
	AmountReceived            *float32                  `json:"amount_received" gorm:"type:decimal(10,2)"`
	DeliveryFee               *float32                  `json:"delivery_fee" gorm:"type:decimal(10,2)"`
	TaxCharge                 float32                   `json:"tax_charge" gorm:"type:decimal(10,2)" default:"0"`
	TaxPercentage             float32                   `json:"tax_percentage" gorm:"type:decimal(10,2)" default:"0"`
	ServiceCharge             float32                   `json:"service_charge" gorm:"type:decimal(10,2)" default:"0"`
	ServiceChargePercentage   float32                   `json:"service_charge_percentage" gorm:"type:decimal(10,2)" default:"0"`
	DiscountType              *string                   `json:"discount_type" gorm:"type:varchar(255)" enum:"fixed,percentage,none"`
	DiscountAmount            float32                   `json:"discount_amount" gorm:"type:decimal(10,2)" default:"0"`
	DiscountPercentage        float32                   `json:"discount_percentage" gorm:"type:decimal(10,2)" default:"0"`
	PaymentMethod             string                    `json:"payment_method" gorm:"type:varchar(255)"`
	PaymentChannel            *PaymentChannel           `json:"payment_channel" gorm:"type:varchar(255)"`
	PaymentStatus             string                    `json:"payment_status" gorm:"type:varchar(255)" enum:"pending,completed,failed,refunded,voided,cancelled" default:"pending"`
	Notes                     string                    `json:"notes" gorm:"type:text"`
	TableNumber               string                    `json:"table_number" gorm:"type:varchar(255)"`
	EInvoiceSubmissionID      *string                   `json:"e_invoice_submission_id" gorm:"type:varchar(255)"`
	EInvoiceStatus            *constants.EInvoiceStatus `json:"e_invoice_status" gorm:"type:varchar(255)"`
	EInvoiceURL               *string                   `json:"e_invoice_url" gorm:"type:varchar(512)"`
	EInvoiceRejectedReason    *string                   `json:"e_invoice_rejected_reason" gorm:"type:varchar(255)"`
	OrderDetails              *OrderDetails             `json:"order_details" gorm:"foreignKey:OrderID"`
	OrderItems                []OrderItem               `json:"order_items" gorm:"foreignKey:OrderID"`
	PickupAt                  *time.Time                `json:"pickup_at" gorm:"type:timestamp with time zone"`          // for schedule pickup later, this time is plan time not actual time
	CompletedAt               *time.Time                `json:"completed_at" gorm:"type:timestamp with time zone"`       // this time is actual time when order is delivered/collected/completed
	PointsRewarded            *int                      `json:"points_rewarded" gorm:"type:integer"`                     // this is total points rewarded to customer based on point rules
	PointsRewardedAt          *time.Time                `json:"points_rewarded_at" gorm:"type:timestamp with time zone"` // this is time when points are rewarded to customer
	VoucherDiscountAmount     *float32                  `json:"voucher_discount_amount" gorm:"type:decimal(10,2)" default:"0"`
	VoucherDiscountType       *VoucherDiscountType      `json:"voucher_discount_type" gorm:"type:varchar(50)" enum:"fixed,percentage,none"`
	VoucherDiscountPercentage *float32                  `json:"voucher_discount_percentage" gorm:"type:decimal(10,2)" default:"0"`
	ExpRewarded               *int                      `json:"exp_rewarded" gorm:"type:integer"`                     // this is total exp rewarded to customer based on point rules
	ExpRewardedAt             *time.Time                `json:"exp_rewarded_at" gorm:"type:timestamp with time zone"` // this is time when exp are rewarded to customer
	CreatedAt                 time.Time                 `json:"created_at"`
	UpdatedAt                 *time.Time                `json:"updated_at"`
	DeletedAt                 gorm.DeletedAt            `json:"deleted_at,omitempty"`
}

type OrderItem struct {
	ID                     uuid.UUID               `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OrderID                uuid.UUID               `json:"order_id" gorm:"type:uuid"`
	Order                  Order                   `json:"order" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	ProductID              uuid.UUID               `json:"product_id" gorm:"type:uuid"`
	Product                Product                 `json:"product" gorm:"foreignKey:ProductID"`
	Quantity               int                     `json:"quantity" gorm:"type:integer"`
	UnitPrice              float32                 `json:"unit_price" gorm:"type:decimal(10,2)"`
	SubTotal               float32                 `json:"sub_total" gorm:"type:decimal(10,2)"`
	ItemNotes              string                  `json:"item_notes" gorm:"type:text"`
	SelectedModifierGroups []SelectedModifierGroup `json:"selected_modifier_groups" gorm:"foreignKey:OrderItemID"`
	CreatedAt              time.Time               `json:"created_at"`
	UpdatedAt              *time.Time              `json:"updated_at"`
	DeletedAt              gorm.DeletedAt          `json:"deleted_at,omitempty"`
}

// use for modifier group and modifier option (latest)
type SelectedModifierGroup struct {
	ID                     uuid.UUID       `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OrderItemID            uuid.UUID       `json:"order_item_id" gorm:"type:uuid;index"`
	OrderItem              OrderItem       `json:"order_item" gorm:"foreignKey:OrderItemID;constraint:OnDelete:CASCADE"`
	ModifierGroupID        uuid.UUID       `json:"modifier_group_id" gorm:"type:uuid;index"`
	ModifierGroup          ModifierGroups  `json:"modifier_group" gorm:"foreignKey:ModifierGroupID"`
	ModifierOptionsID      uuid.UUID       `json:"modifier_options_id" gorm:"type:uuid;index"`
	ModifierOptions        ModifierOptions `json:"modifier_options" gorm:"foreignKey:ModifierOptionsID"`
	ModifierOptionQuantity int             `json:"modifier_option_quantity" gorm:"type:integer"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              *time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt  `json:"deleted_at,omitempty"`
}

type OrderDetails struct {
	ID                      uuid.UUID        `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OrderID                 uuid.UUID        `json:"order_id" gorm:"type:uuid;index"`
	GrabShortOrderNum       *string          `json:"grab_short_order_number" gorm:"type:varchar(20)"`
	ShopeeFoodShortOrderNum *string          `json:"shopeefood_short_order_num" gorm:"type:varchar(20)"`
	CustomerName            *string          `json:"customer_name" gorm:"type:varchar(50)"`
	CustomerPhone           *string          `json:"customer_phone" gorm:"type:varchar(50)"`
	CustomerAddress         *string          `json:"customer_address" gorm:"type:varchar(255)"`
	CustomerLatitude        *float32         `json:"customer_latitude" gorm:"type:decimal(11,8)"`
	CustomerLongitude       *float32         `json:"customer_longitude" gorm:"type:decimal(11,8)"`
	EstimatedOrderReadyTime *time.Time       `json:"estimated_order_ready_time" gorm:"type:timestamp with time zone"`
	MaxOrderReadyTime       *time.Time       `json:"max_order_ready_time" gorm:"type:timestamp with time zone"`
	NewOrderReadyTime       *time.Time       `json:"new_order_ready_time" gorm:"type:timestamp with time zone"`
	VoucherID               *uuid.UUID       `json:"voucher_id" gorm:"type:uuid"`
	Voucher                 *Voucher         `json:"voucher" gorm:"foreignKey:VoucherID"`
	CustomerVoucherID       *uuid.UUID       `json:"customer_voucher_id" gorm:"type:uuid"`
	CustomerVoucher         *CustomerVoucher `json:"customer_voucher" gorm:"foreignKey:CustomerVoucherID"`
	TrackingInfo            *TrackingInfo    `json:"tracking_info" gorm:"embedded"`
	CreatedAt               time.Time        `json:"created_at"`
	UpdatedAt               *time.Time       `json:"updated_at"`
	DeletedAt               gorm.DeletedAt   `json:"deleted_at,omitempty"`
}

type TrackingInfo struct {
	DeliveryID      *string     `json:"deliveryID" gorm:"type:varchar(50)"`
	DeliveryStatus  *string     `json:"delivery_status" gorm:"type:varchar(50)"`
	TrackingURL     *string     `json:"tracking_url" gorm:"type:varchar(512)"`
	FailedReason    *string     `json:"failedReason" gorm:"type:varchar(255)"`
	PickupProofURL  *string     `json:"pickupProofURL" gorm:"type:varchar(512)"`
	DropoffProofURL *string     `json:"dropoffProofURL" gorm:"type:varchar(512)"`
	CancelProofURL  *string     `json:"cancelProofURL" gorm:"type:varchar(512)"`
	DriverInfo      *DriverInfo `json:"driver_info" gorm:"embedded"`
}

type DriverInfo struct {
	DriverName         *string `json:"driver_name" gorm:"type:varchar(50)"`
	DriverPhone        *string `json:"driver_phone" gorm:"type:varchar(20)"`
	DriverLicensePlate *string `json:"driver_license_plate" gorm:"type:varchar(20)"`
	DriverPhotoURL     *string `json:"driver_photo_url" gorm:"type:varchar(512)"`
}
