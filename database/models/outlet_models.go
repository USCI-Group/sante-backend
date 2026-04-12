package models

import (
	"encoding/json"
	"time"

	"encore.app/common"
	"encore.app/common/constants"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type OutletStatus string

const (
	OutletStatusOpen   OutletStatus = "open"
	OutletStatusOnHold OutletStatus = "on-hold"
	OutletStatusClosed OutletStatus = "closed"
)

type DayOfWeek string

const (
	DayOfWeekMonday    DayOfWeek = "monday"
	DayOfWeekTuesday   DayOfWeek = "tuesday"
	DayOfWeekWednesday DayOfWeek = "wednesday"
	DayOfWeekThursday  DayOfWeek = "thursday"
	DayOfWeekFriday    DayOfWeek = "friday"
	DayOfWeekSaturday  DayOfWeek = "saturday"
	DayOfWeekSunday    DayOfWeek = "sunday"
)

type Outlet struct {
	ID                 uuid.UUID                 `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID         uuid.UUID                 `json:"business_id" gorm:"type:uuid" valid:"required~BusinessID is required"`
	Name               string                    `json:"name" gorm:"type:varchar(255)"`
	Email              string                    `json:"email" gorm:"type:varchar(255)"`
	Phone              string                    `json:"phone" gorm:"type:varchar(255)"`
	Address            common.Address            `json:"address" gorm:"embedded"`
	Website            string                    `json:"website,omitempty" gorm:"type:varchar(255)"`
	IDType             constants.IDType          `json:"id_type,omitempty" gorm:"type:varchar(255)"`
	RegistrationNumber string                    `json:"registration_number,omitempty" gorm:"type:varchar(255)"`
	TIN                *string                   `json:"tin,omitempty" gorm:"type:varchar(255)"`
	ImageURL           string                    `json:"image_url,omitempty" gorm:"type:varchar(255)"`
	Business           *Business                 `json:"business" gorm:"foreignKey:BusinessID"`
	OutletStatus       OutletStatus              `json:"outlet_status" gorm:"type:varchar(10);default:closed"`
	OutletStaticQR     *string                   `json:"outlet_static_qr,omitempty" gorm:"type:varchar(255)"`
	OutletGroups       []*OutletGroup            `gorm:"many2many:outlet_groups_outlets;"`
	Latitude           *float64                  `json:"latitude,omitempty" gorm:"type:decimal(12,8)"`
	Longitude          *float64                  `json:"longitude,omitempty" gorm:"type:decimal(12,8)"`
	OperationSchedules []OutletOperationSchedule `json:"operation_schedules" gorm:"foreignKey:OutletID"`
	OperationTimeSlots []OutletOperationTimeSlot `json:"operation_time_slots" gorm:"foreignKey:OutletID"`
	OnlineOrderEnabled bool                      `json:"online_order_enabled" gorm:"type:boolean; default:true"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          *time.Time                `json:"updated_at"`
	DeletedAt          gorm.DeletedAt            `json:"deleted_at,omitempty"`
}

type OutletGroup struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"type:varchar(255)"`
	Description string         `json:"description" gorm:"type:varchar(255)"`
	BusinessID  uuid.UUID      `json:"business_id" gorm:"type:uuid"`
	Business    Business       `json:"business" gorm:"foreignKey:BusinessID"`
	Outlets     []*Outlet      `gorm:"many2many:outlet_groups_outlets;"`
	Users       []*User        `gorm:"many2many:outlet_groups_users;"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type MerchantSecret struct {
	ID                          uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OutletID                    uuid.UUID      `json:"outlet_id" gorm:"type:uuid;unique"`
	FiuuMerchantID              *string        `json:"fiuu_merchant_id" gorm:"type:varchar(255)"`
	FiuuVerifyKey               *string        `json:"fiuu_verify_key" gorm:"type:varchar(255)"`                 // For Online Payment (Encrypted)
	FiuuSecretKey               *string        `json:"fiuu_secret_key" gorm:"type:varchar(255)"`                 // For Online Payment (Encrypted)
	FiuuApplicationCode         *string        `json:"fiuu_application_code" gorm:"type:varchar(255)"`           // For Offline Payment
	FiuuOfflineSecretKey        *string        `json:"fiuu_offline_secret_key" gorm:"type:varchar(255)"`         // For Offline Payment (Encrypted)
	FiuuCloudERCAccountID       *string        `json:"fiuu_cloud_erc_account_id" gorm:"type:varchar(255)"`       // For Cloud ECR
	FiuuCloudERCAccountPassword *string        `json:"fiuu_cloud_erc_account_password" gorm:"type:varchar(255)"` // For Cloud ECR (Encrypted)
	FiuuCloudERCSecretKey       *string        `json:"fiuu_cloud_erc_secret_key" gorm:"type:varchar(255)"`       // For Cloud ECR (Encrypted)
	EInvoiceAuthToken           *string        `json:"e_invoice_auth_token" gorm:"type:varchar(2048)"`
	EInvoiceAuthTokenExpiry     *time.Time     `json:"e_invoice_auth_token_expiry" gorm:"type:timestamp with time zone"`
	GrabStoreID                 *string        `json:"grab_store_id" gorm:"type:varchar(255)"`
	ShopeeStoreID               *string        `json:"shopee_store_id" gorm:"type:varchar(255)"`
	GrabIntegrationStatus       *string        `json:"grab_integration_status" gorm:"type:varchar(255)"`
	ShopeeIntegrationStatus     *string        `json:"shopee_integration_status" gorm:"type:varchar(255)"`
	GrabMenuSyncState           *string        `json:"grab_menu_sync_state" gorm:"type:varchar(50)"`
	ShopeeMenuSyncState         *string        `json:"shopee_menu_sync_state" gorm:"type:varchar(50)"`
	CreatedAt                   time.Time      `json:"created_at"`
	UpdatedAt                   *time.Time     `json:"updated_at"`
	DeletedAt                   gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type OutletProduct struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OutletID  uuid.UUID      `json:"outlet_id" gorm:"type:uuid"`
	Outlet    Outlet         `json:"outlet" gorm:"foreignKey:OutletID"`
	ProductID uuid.UUID      `json:"product_id" gorm:"type:uuid"`
	Product   Product        `json:"product" gorm:"foreignKey:ProductID"`
	IsActive  bool           `json:"is_active" gorm:"type:boolean; default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type OutletModifierOption struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OutletID          uuid.UUID       `json:"outlet_id" gorm:"type:uuid"`
	Outlet            Outlet          `json:"outlet" gorm:"foreignKey:OutletID"`
	ModifierOptionsID uuid.UUID       `json:"modifier_options_id" gorm:"type:uuid;index"`
	ModifierOptions   ModifierOptions `json:"modifier_options" gorm:"foreignKey:ModifierOptionsID"`
	IsActive          bool            `json:"is_active" gorm:"type:boolean; default:true"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         *time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt  `json:"deleted_at,omitempty"`
}

type OutletOperationSchedule struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OutletID       uuid.UUID      `json:"outlet_id" gorm:"type:uuid;index:idx_outlet_schedule"`
	Outlet         *Outlet        `json:"outlet,omitempty" gorm:"foreignKey:OutletID"`
	DayOfWeek      DayOfWeek      `json:"day_of_week" gorm:"type:varchar(255);index:idx_outlet_schedule"`
	IsClosed       bool           `json:"is_closed" gorm:"type:boolean; default:false"`
	OpenTime       time.Time      `json:"open_time" gorm:"type:timestamp with time zone"`
	CloseTime      time.Time      `json:"close_time" gorm:"type:timestamp with time zone"`
	BreakStartTime *time.Time     `json:"break_start_time,omitempty" gorm:"type:timestamp with time zone"`
	BreakEndTime   *time.Time     `json:"break_end_time,omitempty" gorm:"type:timestamp with time zone"`
	IsBreakActive  bool           `json:"is_break_active" gorm:"type:boolean; default:false"`
	IsActive       bool           `json:"is_active" gorm:"type:boolean; default:true;index"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      *time.Time     `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type OutletOperationTimeSlot struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OutletID          uuid.UUID      `json:"outlet_id" gorm:"type:uuid;index:idx_pickup_slot"`
	Outlet            *Outlet        `json:"outlet,omitempty" gorm:"foreignKey:OutletID"`
	StartTime         time.Time      `json:"start_time" gorm:"type:timestamp with time zone;index:idx_pickup_slot"`
	EndTime           time.Time      `json:"end_time" gorm:"type:timestamp with time zone;index:idx_pickup_slot"`
	IsPickupAvailable bool           `json:"is_pickup_available" gorm:"type:boolean; default:true;index"` // true mean available for pickup
	IsActive          bool           `json:"is_active" gorm:"type:boolean; default:true;index"`           // this is overall active or not
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         *time.Time     `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type OutletTerminal struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OutletID  uuid.UUID  `json:"outlet_id" gorm:"type:uuid;index:idx_outlet_terminal"`
	Outlet    *Outlet    `json:"outlet,omitempty" gorm:"foreignKey:OutletID;constraint:OnDelete:CASCADE"`
	VtID      string     `json:"vt_id" gorm:"type:varchar(255);uniqueIndex"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// MarshalJSON customizes JSON output to show only time portion
func (o OutletOperationSchedule) MarshalJSON() ([]byte, error) {
	type Alias OutletOperationSchedule
	return json.Marshal(&struct {
		OpenTime       string  `json:"open_time"`
		CloseTime      string  `json:"close_time"`
		BreakStartTime *string `json:"break_start_time,omitempty"`
		BreakEndTime   *string `json:"break_end_time,omitempty"`
		*Alias
	}{
		OpenTime:       o.OpenTime.Format("15:04:05"),
		CloseTime:      o.CloseTime.Format("15:04:05"),
		BreakStartTime: formatTimePtr(o.BreakStartTime),
		BreakEndTime:   formatTimePtr(o.BreakEndTime),
		Alias:          (*Alias)(&o),
	})
}

// MarshalJSON customizes JSON output to show only time portion
func (o OutletOperationTimeSlot) MarshalJSON() ([]byte, error) {
	type Alias OutletOperationTimeSlot
	return json.Marshal(&struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		*Alias
	}{
		StartTime: o.StartTime.Format("15:04:05"),
		EndTime:   o.EndTime.Format("15:04:05"),
		Alias:     (*Alias)(&o),
	})
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("15:04:05")
	return &s
}
