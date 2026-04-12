package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type MissionRewardType string

const (
	MissionRewardTypePoints  MissionRewardType = "points"
	MissionRewardTypeVoucher MissionRewardType = "voucher"
)

type MissionCriteriaType string

const (
	MissionCriteriaTypeOrderCount     MissionCriteriaType = "order_count"
	MissionCriteriaTypeProductCount   MissionCriteriaType = "product_count"
	MissionCriteriaTypeSpendingAmount MissionCriteriaType = "spending_amount"
	MissionCriteriaTypeLocationBased  MissionCriteriaType = "location_based"
	MissionCriteriaTypeMembership     MissionCriteriaType = "membership"
)

type MissionFrequency string

const (
	MissionFrequencyOneTime MissionFrequency = "one-time"
	MissionFrequencyDaily   MissionFrequency = "daily"
)

type Mission struct {
	ID                 uuid.UUID         `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID         uuid.UUID         `json:"business_id" gorm:"type:uuid"`
	Name               string            `json:"name" gorm:"type:varchar(255)"`
	Description        string            `json:"description" gorm:"type:text"`
	Cost               int               `json:"cost" gorm:"type:int"`                            // cost to join mission                       // reward for completing mission
	IsActive           bool              `json:"is_active" gorm:"type:boolean; default:false"`    // is active mission
	ValidFrom          time.Time         `json:"valid_from" gorm:"type:timestamp with time zone"` // valid from
	ValidTo            time.Time         `json:"valid_to" gorm:"type:timestamp with time zone"`   // valid to
	Validity           int               `json:"validity" gorm:"type:int"`
	Frequency          MissionFrequency  `json:"frequency" gorm:"type:varchar(50)"`
	MissionCriteria    []MissionCriteria `json:"criteria" gorm:"foreignKey:MissionID"`
	MissionRewards     []MissionReward   `json:"mission_rewards" gorm:"foreignKey:MissionID"` // mission rewards (points, vouchers)
	MembershipID       *uuid.UUID        `json:"membership_id,omitempty" gorm:"type:uuid"`    // eligible membership id for this mission
	ImageURL           string            `json:"image_url" gorm:"type:varchar(255)"`
	TermsAndConditions *string           `json:"terms_and_conditions" gorm:"type:text"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          *time.Time        `json:"updated_at"`
	DeletedAt          gorm.DeletedAt    `json:"deleted_at,omitempty"`
}

// MissionCriteria is a criteria to complete the mission
// Each row represents exactly ONE type of criteria (products, membership, outlet, or field-based)
// location based criteria although is exisit as a new row of data, but it purpose is to use for product / order.
// Ensure only product / order purchased in the location will be counted.
// membership filled is serve as a minimum requirement rank. Therefore, if user tier is higher than the mission tier, it will be counted.
type MissionCriteria struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	MissionID uuid.UUID `json:"mission_id" gorm:"type:uuid"`

	// Discriminator field - determines which criteria type this row represents
	CriteriaType MissionCriteriaType `json:"criteria_type" gorm:"type:varchar(50);not null"`

	// Only ONE of these should be populated based on CriteriaType
	// For product-based criteria
	ProductID *uuid.UUID `json:"product_id,omitempty" gorm:"type:uuid"`
	Product   *Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	// For membership-based criteria
	MembershipID *uuid.UUID  `json:"membership_id,omitempty" gorm:"type:uuid"`
	Membership   *Membership `json:"membership,omitempty" gorm:"foreignKey:MembershipID"`

	// For outlet-based criteria
	OutletID *uuid.UUID `json:"outlet_id,omitempty" gorm:"type:uuid"`
	Outlet   *Outlet    `json:"outlet,omitempty" gorm:"foreignKey:OutletID"`

	// For field-based criteria (order count, spending amount)
	Value float64 `json:"value,omitempty" gorm:"type:numeric(10,2);default:0"`

	IsActive  bool           `json:"is_active" gorm:"type:boolean;default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type MissionReward struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	MissionID uuid.UUID `json:"mission_id" gorm:"type:uuid"`
	//Mission     Mission           `json:"mission" gorm:"foreignKey:MissionID"`
	RewardType  MissionRewardType `json:"reward_type" gorm:"type:varchar(50);check:reward_type IN ('points', 'voucher')"`
	PointRuleID *uuid.UUID        `json:"point_rule_id,omitempty" gorm:"type:uuid"` // voucher and point rule cannot exist at the same row
	PointRule   *PointRule        `json:"point_rule,omitempty" gorm:"foreignKey:PointRuleID"`
	VoucherID   *uuid.UUID        `json:"voucher_id,omitempty" gorm:"type:uuid"` // voucher and point rule cannot exist at the same row
	Voucher     *Voucher          `json:"voucher,omitempty" gorm:"foreignKey:VoucherID"`
	Quantity    int               `json:"quantity" gorm:"type:int;default:1"` // How many rewards to give
	IsActive    bool              `json:"is_active" gorm:"type:boolean;default:true"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   *time.Time        `json:"updated_at"`
	DeletedAt   gorm.DeletedAt    `json:"deleted_at,omitempty"`
}
