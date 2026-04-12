package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type VoucherType string
type VoucherFor string
type VoucherEligibleOrderMethod string
type VoucherEligibleUserType string
type VoucherPlatform string
type VoucherEligibilityRuleType string

const (
	VoucherDiscount         VoucherType = "discount"
	VoucherGiveaway         VoucherType = "giveaway"
	VoucherPointsRedemption VoucherType = "points redemption"
	VoucherPointsEarnings   VoucherType = "points earnings"
	VoucherOther            VoucherType = "other"
)

const (
	VoucherForGeneral  VoucherFor = "general"
	VoucherForBirthday VoucherFor = "birthday"
	VoucherForRanking  VoucherFor = "ranking"
	VoucherForReferral VoucherFor = "referral"
	VoucherForLoyalty  VoucherFor = "loyalty"
	VoucherForLTO      VoucherFor = "lto"
)

const (
	VoucherEligibleOrderMethodPickup            VoucherEligibleOrderMethod = "pickup"
	VoucherEligibleOrderMethodDelivery          VoucherEligibleOrderMethod = "delivery"
	VoucherEligibleOrderMethodPickupAndDelivery VoucherEligibleOrderMethod = "pickup_and_delivery"
	VoucherEligibleOrderMethodAll               VoucherEligibleOrderMethod = "all"
)

const (
	VoucherEligibleUserTypeNewCustomer      VoucherEligibleUserType = "new_customer"
	VoucherEligibleUserTypeExistingCustomer VoucherEligibleUserType = "existing_customer"
	VoucherEligibleUserTypeAll              VoucherEligibleUserType = "all"
)

const (
	VoucherPlatformWeb              VoucherPlatform = "web"
	VoucherPlatformMobilePOS        VoucherPlatform = "mobile_pos"
	VoucherPlatformMobileMembership VoucherPlatform = "mobile_membership"
	VoucherPlatformAll              VoucherPlatform = "all"
)

const (
	VoucherEligibilityRuleOutlet          VoucherEligibilityRuleType = "outlet"
	VoucherEligibilityRuleMembership      VoucherEligibilityRuleType = "membership"
	VoucherEligibilityRuleUser            VoucherEligibilityRuleType = "user"
	VoucherEligibilityRuleProduct         VoucherEligibilityRuleType = "product"
	VoucherEligibilityRuleProductCategory VoucherEligibilityRuleType = "product_category"
)

// Voucher represents a discount or promotion voucher.
type Voucher struct {
	ID                        uuid.UUID                  `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID                *uuid.UUID                 `json:"business_id" gorm:"type:uuid"`
	Name                      string                     `json:"name" gorm:"type:varchar(255)"`
	Description               string                     `json:"description" gorm:"type:text"`
	TermsAndConditions        *string                    `json:"terms_and_conditions" gorm:"type:text"`
	VoucherCode               string                     `json:"voucher_code" gorm:"type:varchar(255);uniqueIndex;not null"`
	VoucherImageURL           *string                    `json:"voucher_image_url" gorm:"type:text"`
	VoucherFor                VoucherFor                 `json:"voucher_for" gorm:"type:varchar(255)"`
	VoucherType               VoucherType                `json:"voucher_type" gorm:"type:varchar(255)"`
	MinPurchase               *float32                   `json:"min_purchase" gorm:"type:numeric"`
	MaxPurchase               *float32                   `json:"max_purchase" gorm:"type:numeric"`
	MaxRedemption             *int                       `json:"max_redemption" gorm:"type:int"`
	MaxRedemptionPerCustomer  *int                       `json:"max_redemption_per_customer" gorm:"type:int"`
	CurrentRedemptions        int                        `json:"current_redemptions" gorm:"type:int;default:0"`    // THIS IS FOR TOTAL REDEMPTIONS OF VOUCHER
	CurrentUsage              int                        `json:"current_usage" gorm:"type:int;default:0"`          // THIS IS FOR USAGE OF VOUCHER AFTER EACH PAYMENT
	RedeemValue               float32                    `json:"redeem_value" gorm:"type:decimal(10,2);default:0"` // THIS IS FOR REDEEM VALUE OF VOUCHER
	IsActive                  bool                       `json:"is_active" gorm:"type:boolean;default:true"`
	DiscountID                *uuid.UUID                 `json:"discount_id" gorm:"type:uuid"`
	Discount                  *Discount                  `json:"discount" gorm:"foreignKey:DiscountID"`
	DiscountType              *DiscountType              `json:"discount_type" gorm:"type:varchar(50)"`
	DiscountValue             *float32                   `json:"discount_value" gorm:"type:decimal(10,2)"`
	IsEligibleForRankingClimb bool                       `json:"is_eligible_for_ranking_climb" gorm:"type:boolean;default:false"`
	EligibleOrderMethod       VoucherEligibleOrderMethod `json:"eligible_order_method" gorm:"type:varchar(50)"`
	EligiblePlatform          VoucherPlatform            `json:"eligible_platform" gorm:"type:varchar(50)"`
	EligibleUserType          VoucherEligibleUserType    `json:"eligible_user_type" gorm:"type:varchar(50)"`
	IsStackable               bool                       `json:"is_stackable" gorm:"type:boolean;default:false"`
	IsExclusive               bool                       `json:"is_exclusive" gorm:"type:boolean;default:true"`
	IsOneTimeUse              bool                       `json:"is_one_time_use" gorm:"type:boolean;default:true"`
	IsMobileAppOnly           bool                       `json:"is_mobile_app_only" gorm:"type:boolean;default:false"`
	Priority                  int                        `json:"priority" gorm:"type:int;default:0"`
	ValidFrom                 time.Time                  `json:"valid_from" gorm:"type:timestamp with time zone"`
	ValidTo                   time.Time                  `json:"valid_to" gorm:"type:timestamp with time zone"`
	Validity                  *int                       `json:"validity" gorm:"type:int"` // THIS IS FOR VALIDITY OF VOUCHER IN DAYS
	CreatedBy                 *uuid.UUID                 `json:"created_by" gorm:"type:uuid"`
	UpdatedBy                 *uuid.UUID                 `json:"updated_by" gorm:"type:uuid"`
	VoucherEligibilityRules   []VoucherEligibilityRule   `json:"voucher_eligibility_rules" gorm:"foreignKey:VoucherID"`
	ProductID                 *uuid.UUID                 `json:"product_id" gorm:"type:uuid"` // THIS IS FOR GIVEAWAY VOUCHER TYPE ONLY
	Product                   *Product                   `json:"product" gorm:"foreignKey:ProductID"`
	GiveawayAmount            *int                       `json:"giveaway_amount" gorm:"type:int"`
	CreatedAt                 time.Time                  `json:"created_at"`
	UpdatedAt                 *time.Time                 `json:"updated_at"`
	DeletedAt                 gorm.DeletedAt             `json:"deleted_at,omitempty"`
}

// VoucherOutlet join table for voucher and outlets
type VoucherOutlet struct {
	// composite primary key
	VoucherID uuid.UUID      `gorm:"type:uuid;primaryKey"`
	OutletID  uuid.UUID      `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type VoucherEligibilityRule struct {
	ID                  uuid.UUID                  `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	VoucherID           uuid.UUID                  `json:"voucher_id" gorm:"type:uuid"`
	EligibilityRuleType VoucherEligibilityRuleType `json:"eligibility_rule_type" gorm:"type:varchar(255)"`
	OutletID            *uuid.UUID                 `json:"outlet_id" gorm:"type:uuid"`
	Outlet              *Outlet                    `json:"outlet,omitempty" gorm:"foreignKey:OutletID"`
	ProductID           *uuid.UUID                 `json:"product_id" gorm:"type:uuid"`
	Product             *Product                   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	ProductCategoryID   *uuid.UUID                 `json:"product_category_id" gorm:"type:uuid"`
	ProductCategory     *ProductCategory           `json:"product_category,omitempty" gorm:"foreignKey:ProductCategoryID"`
	MembershipID        *uuid.UUID                 `json:"membership_id" gorm:"type:uuid"`
	Membership          *Membership                `json:"membership,omitempty" gorm:"foreignKey:MembershipID"`
	CreatedAt           time.Time                  `json:"created_at"`
	UpdatedAt           *time.Time                 `json:"updated_at"`
}
