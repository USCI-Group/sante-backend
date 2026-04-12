package models

import (
	"time"

	"encore.app/common"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type PaymentMethod string

const (
	PaymentMethodFPX          PaymentMethod = "fpx"
	PaymentMethodBNPL         PaymentMethod = "bnpl"
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodCard         PaymentMethod = "card"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodDebitCard    PaymentMethod = "debit_card"
	PaymentMethodEWallet      PaymentMethod = "e-wallet"
	PaymentMethodStaticQR     PaymentMethod = "static_qr"
	PaymentMethodGrabCash     PaymentMethod = "CASH"
	PaymentMethodGrabCashless PaymentMethod = "CASHLESS"
	PaymentMethodOther        PaymentMethod = "other"
)

type PaymentChannel string

const (
	// E-Wallet Payment Channels
	PaymentChannelTouchNGoEWallet PaymentChannel = "Touch 'n Go eWallet"
	PaymentChannelBoost           PaymentChannel = "Boost"
	PaymentChannelShopeePay       PaymentChannel = "ShopeePay"
	PaymentChannelEWallet         PaymentChannel = "E-wallet"

	// Online Banking Payment Channels
	PaymentChannelMaybank2u        PaymentChannel = "Maybank2u"
	PaymentChannelCIMBClicks       PaymentChannel = "CIMB Clicks"
	PaymentChannelPublicBankOnline PaymentChannel = "Public Bank Online"
	PaymentChannelHongLeongConnect PaymentChannel = "Hong Leong Connect"
	PaymentChannelRHBNow           PaymentChannel = "RHB Now"
	PaymentChannelAmOnline         PaymentChannel = "AmOnline"
	PaymentChannelUOBPIB           PaymentChannel = "UOB Personal Internet Banking"

	// Card Payment Channels
	PaymentChannelCard       PaymentChannel = "Card"
	PaymentChannelVisa       PaymentChannel = "Visa"
	PaymentChannelMastercard PaymentChannel = "Mastercard"

	//cash
	PaymentChannelCash PaymentChannel = "Cash"
)

type PaymentPlatform string

const (
	PaymentPlatformPOS           PaymentPlatform = "pos"
	PaymentPlatformMembershipApp PaymentPlatform = "membership_app"
	PaymentPlatformWeb           PaymentPlatform = "web"
)

type ComplianceStatus string

const (
	ComplianceStatusActive      ComplianceStatus = "active"
	ComplianceStatusSuspended   ComplianceStatus = "suspended"
	ComplianceStatusUnderReview ComplianceStatus = "under_review"
)

type Business struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	Name               string         `json:"name" gorm:"type:varchar(255)"`
	Email              string         `json:"email" gorm:"type:varchar(255)"`
	Phone              string         `json:"phone" gorm:"type:varchar(255)"`
	Address            common.Address `json:"address" gorm:"embedded"`
	Website            string         `json:"website" gorm:"type:varchar(255)"`
	RegistrationNumber string         `json:"registration_number" gorm:"type:varchar(255)"`
	LogoURL            string         `json:"logo_url" gorm:"type:varchar(255)"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          *time.Time     `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type BusinessConfiguration struct {
	ID                                  uuid.UUID                `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID                          uuid.UUID                `json:"business_id" gorm:"type:uuid"`
	Business                            *Business                `json:"business,omitempty" gorm:"foreignKey:BusinessID"`
	GrabClientID                        *string                  `json:"grab_client_id" gorm:"type:varchar(255)"`             // GrabFood Client ID
	GrabClientSecret                    *string                  `json:"grab_client_secret" gorm:"type:varchar(255)"`         // encrypted
	GrabExpressClientID                 *string                  `json:"grab_express_client_id" gorm:"type:varchar(255)"`     // GrabExpress Client ID
	GrabExpressClientSecret             *string                  `json:"grab_express_client_secret" gorm:"type:varchar(255)"` // encrypted
	ShopeeClientID                      *string                  `json:"shopee_client_id" gorm:"type:varchar(255)"`
	ShopeeClientSecret                  *string                  `json:"shopee_client_secret" gorm:"type:varchar(255)"` // encrypted
	ServiceChargePercentage             *float32                 `json:"service_charge_percentage" gorm:"type:decimal(5,2); default:0.00"`
	ServiceTaxPercentage                *float32                 `json:"service_tax_percentage" gorm:"type:decimal(5,2); default:0.00"`
	IsTaxIncludedInPrice                bool                     `json:"is_tax_included_in_price" gorm:"type:boolean; default:true"`
	IsLogoutButtonVisible               bool                     `json:"is_logout_button_visible" gorm:"type:boolean; default:false"`
	TermsOfService                      *string                  `json:"terms_of_service" gorm:"type:text"`
	PrivacyPolicy                       *string                  `json:"privacy_policy" gorm:"type:text"`
	MembershipUpgradeMethod             *MembershipUpgradeMethod `json:"membership_upgrade_method" gorm:"type:varchar(255)"`
	MembershipReviewPeriod              *MembershipReviewPeriod  `json:"membership_review_period" gorm:"type:varchar(50); default:'monthly'"`
	ExperiencePointsPerCurrency         *int                     `json:"experience_points_per_currency" gorm:"type:integer;"`
	IsAutoCompleteOrderOnPayment        *bool                    `json:"is_auto_complete_order_on_payment,omitempty" gorm:"type:boolean; default:false"`
	GoogleAuthClientID                  *string                  `json:"google_auth_client_id" gorm:"type:varchar(255)"`
	EnableInputStockOpening             *bool                    `json:"enable_input_stock_opening" gorm:"type:boolean; default:true"`
	EnableInputCashOpening              *bool                    `json:"enable_input_cash_opening" gorm:"type:boolean; default:true"`
	EnableInputStockClosing             *bool                    `json:"enable_input_stock_closing" gorm:"type:boolean; default:true"`
	EnableInputIngredientWastageClosing *bool                    `json:"enable_input_ingredient_wastage_closing" gorm:"type:boolean; default:true"`
	EnableInputProductWastageClosing    *bool                    `json:"enable_input_product_wastage_closing" gorm:"type:boolean; default:true"`
	EnableInputCashClosing              *bool                    `json:"enable_input_cash_closing" gorm:"type:boolean; default:true"`
	EnableInputExpensesClosing          *bool                    `json:"enable_input_expenses_closing" gorm:"type:boolean; default:true"`
	EnableTerminalMode                  *bool                    `json:"enable_terminal_mode" gorm:"type:boolean; default:false"`                   // if true, the frontend will allow to switch to terminal mode
	EnableTakeAwayPOS                   *bool                    `json:"enable_take_away_pos" gorm:"type:boolean; default:true"`                    // if true, the POS will allow to use take away mode
	EnablePickupPOS                     *bool                    `json:"enable_pickup_pos" gorm:"type:boolean; default:true"`                       // if true, the POS will allow to use pickup mode & pickup later mode
	EnableDineInPOS                     *bool                    `json:"enable_dine_in_pos" gorm:"type:boolean; default:true"`                      // if true, the POS will allow to use dine in mode
	AutoCompleteOrderOnPaymentSuccess   *bool                    `json:"auto_complete_order_on_payment_success" gorm:"type:boolean; default:false"` // if true, the order will be automatically completed on successful payment
	CreatedAt                           time.Time                `json:"created_at"`
	UpdatedAt                           *time.Time               `json:"updated_at"`
	DeletedAt                           gorm.DeletedAt           `json:"deleted_at,omitempty"`
}

type PaymentMethodConfiguration struct {
	ID                 uuid.UUID       `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID         uuid.UUID       `json:"business_id" gorm:"type:uuid"`
	Business           *Business       `json:"business,omitempty" gorm:"foreignKey:BusinessID"`
	PaymentMethod      PaymentMethod   `json:"payment_method" gorm:"type:varchar(255)"`
	PaymentChannel     PaymentChannel  `json:"payment_channel" gorm:"type:varchar(255)"`
	PaymentPlatform    PaymentPlatform `json:"payment_platform" gorm:"type:varchar(255)"`
	PaymentChannelCode *string         `json:"payment_channel_code" gorm:"type:varchar(50)"`      // some payment channel like FPX has multiple codes, so we need to store the code here
	IsActive           bool            `json:"is_active" gorm:"type:boolean; default:true"`       // the hierachy level isActive -> isMaintenance -> isVisible
	IsMaintenance      bool            `json:"is_maintenance" gorm:"type:boolean; default:false"` // the hierachy level isActive -> isMaintenance -> isVisible
	IsVisible          bool            `json:"is_visible" gorm:"type:boolean; default:true"`      // the hierachy level isActive -> isMaintenance -> isVisible

	// Add validation fields
	MinAmount         *float32 `json:"min_amount" gorm:"type:decimal(10,2)"`           // Minimum transaction amount
	MaxAmount         *float32 `json:"max_amount" gorm:"type:decimal(10,2)"`           // Maximum transaction amount
	ProcessingFee     *float32 `json:"processing_fee" gorm:"type:decimal(10,2)"`       // Fixed processing fee
	ProcessingFeeRate *float32 `json:"processing_fee_rate" gorm:"type:decimal(5,4)"`   // Percentage processing fee (0.025 = 2.5%)
	Currency          string   `json:"currency" gorm:"type:varchar(3); default:'MYR'"` // Currency code
	SettlementDays    *int     `json:"settlement_days" gorm:"type:integer"`            // Days to settlement
	Priority          int      `json:"priority" gorm:"type:integer; default:0"`        // Display priority (higher = shown first)

	// Configuration metadata
	DisplayName string  `json:"display_name" gorm:"type:varchar(255)"` // Custom display name
	Description *string `json:"description" gorm:"type:text"`          // Description for users
	IconURL     *string `json:"icon_url" gorm:"type:varchar(500)"`     // Payment method icon
	ColorCode   *string `json:"color_code" gorm:"type:varchar(7)"`     // Hex color for UI

	// Compliance fields
	ComplianceStatus ComplianceStatus `json:"compliance_status" gorm:"type:varchar(50); default:'active'"` // active, suspended, under_review
	ComplianceNotes  *string          `json:"compliance_notes" gorm:"type:text"`
	ValidFrom        *time.Time       `json:"valid_from"`
	ValidUntil       *time.Time       `json:"valid_until"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}
