package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type MembershipUpgradeMethod string

const (
	MembershipUpgradeMethodNoOfProducts        MembershipUpgradeMethod = "no_of_products"
	MembershipUpgradeMethodPriceToExperience   MembershipUpgradeMethod = "price_to_experience"
	MembershipUpgradeMethodProductToExperience MembershipUpgradeMethod = "product_to_experience"
)

type MembershipBenefitLinkType string

const (
	MembershipBenefitLinkTypeProduct  MembershipBenefitLinkType = "product"
	MembershipBenefitLinkTypeVoucher  MembershipBenefitLinkType = "voucher"
	MembershipBenefitLinkTypePoints   MembershipBenefitLinkType = "points"
	MembershipBenefitLinkTypeDiscount MembershipBenefitLinkType = "discount"
)

type MembershipReviewPeriod string

const (
	MembershipReviewPeriodMonthly    MembershipReviewPeriod = "monthly"
	MembershipReviewPeriodQuarterly  MembershipReviewPeriod = "quarterly"
	MembershipReviewPeriodHalfYearly MembershipReviewPeriod = "half_yearly"
	MembershipReviewPeriodAnnually   MembershipReviewPeriod = "annually"
)

type MembershipBenefitType string

const (
	MembershipBenefitTypeDiscount     MembershipBenefitType = "discount"
	MembershipBenefitTypeFreeDelivery MembershipBenefitType = "free_delivery"
	MembershipBenefitTypeFreeVoucher  MembershipBenefitType = "free_voucher"
	MembershipBenefitTypeFreeProduct  MembershipBenefitType = "free_product"
	MembershipBenefitTypeFreePoints   MembershipBenefitType = "free_points"
	MembershipBenefitTypeEarnPoints   MembershipBenefitType = "earn_points"
)

type Membership struct {
	ID           uuid.UUID                `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID   uuid.UUID                `json:"business_id" gorm:"type:uuid"`
	TierName     string                   `json:"tier_name" gorm:"type:varchar(255)"`
	TierLevel    int                      `json:"tier_level" gorm:"type:int;default:0"`
	TierImage    string                   `json:"tier_image" gorm:"type:varchar(255)"`
	Benefits     *[]MembershipBenefit     `json:"benefits,omitempty" gorm:"foreignKey:MembershipID;nullable"`
	UpgradeRules *[]MembershipUpgradeRule `json:"upgrade_rules,omitempty" gorm:"foreignKey:MembershipID;nullable"` // The max criteria for this tier. (To upgrade to this tier)
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    *time.Time               `json:"updated_at"`
	DeletedAt    gorm.DeletedAt           `json:"deleted_at,omitempty"`
}

type MembershipBenefit struct {
	ID                 uuid.UUID             `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	MembershipID       uuid.UUID             `json:"membership_id" gorm:"type:uuid"`
	BenefitName        string                `json:"benefit_name" gorm:"type:varchar(255)"`
	BenefitDescription string                `json:"benefit_description" gorm:"type:text"`
	BenefitType        MembershipBenefitType `json:"benefit_type" gorm:"type:varchar(255)"`
	BenefitValue       string                `json:"benefit_value" gorm:"type:varchar(255)"`
	BenefitImage       string                `json:"benefit_image" gorm:"type:varchar(255)"`
	PointRuleID        *uuid.UUID            `json:"point_rule_id,omitempty" gorm:"type:uuid"`
	PointRule          *PointRule            `json:"point_rule,omitempty" gorm:"foreignKey:PointRuleID"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          *time.Time            `json:"updated_at"`
	DeletedAt          gorm.DeletedAt        `json:"deleted_at,omitempty"`
}

type MembershipUpgradeRule struct {
	ID           uuid.UUID               `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	MembershipID uuid.UUID               `json:"membership_id" gorm:"type:uuid"`
	Method       MembershipUpgradeMethod `json:"method" gorm:"type:varchar(255)"`
	RuleName     string                  `json:"rule_name" gorm:"type:varchar(255)"`
	RuleValue    float32                 `json:"rule_value" gorm:"type:decimal(10,2)"`
	ReviewPeriod MembershipReviewPeriod  `json:"review_period" gorm:"type:varchar(50)"`
	ProductID    *uuid.UUID              `json:"product_id,omitempty" gorm:"type:uuid"`
	Product      *Product                `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    *time.Time              `json:"updated_at"`
	DeletedAt    gorm.DeletedAt          `json:"deleted_at,omitempty"`
}

// Enhanced membership statistics record table - FOCUSED ON METRICS ONLY
type CustomerMembershipStats struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	CustomerID uuid.UUID `json:"customer_id" gorm:"type:uuid;index"`
	BusinessID uuid.UUID `json:"business_id" gorm:"type:uuid;index"`

	// Experience-based metrics
	TotalExperiencePoints int `json:"total_experience_points" gorm:"type:int;default:0"`

	// Spending-based metrics
	TotalSpendingAmount      float32 `json:"total_spending_amount" gorm:"type:decimal(10,2);default:0"`
	MonthlySpendingAmount    float32 `json:"monthly_spending_amount" gorm:"type:decimal(10,2);default:0"`
	QuarterlySpendingAmount  float32 `json:"quarterly_spending_amount" gorm:"type:decimal(10,2);default:0"`
	HalfYearlySpendingAmount float32 `json:"half_yearly_spending_amount" gorm:"type:decimal(10,2);default:0"`
	YearlySpendingAmount     float32 `json:"yearly_spending_amount" gorm:"type:decimal(10,2);default:0"`

	// Order-based metrics
	TotalOrdersCount      int `json:"total_orders_count" gorm:"type:int;default:0"`
	MonthlyOrdersCount    int `json:"monthly_orders_count" gorm:"type:int;default:0"`
	QuarterlyOrdersCount  int `json:"quarterly_orders_count" gorm:"type:int;default:0"`
	HalfYearlyOrdersCount int `json:"half_yearly_orders_count" gorm:"type:int;default:0"`
	YearlyOrdersCount     int `json:"yearly_orders_count" gorm:"type:int;default:0"`

	// Product-based metrics (for product purchase rules)
	ProductPurchaseStats *[]CustomerProductPurchaseStats `json:"product_purchase_stats,omitempty" gorm:"foreignKey:CustomerMembershipStatsID;nullable"`

	// Review period tracking
	LastReviewDate *time.Time `json:"last_review_date" gorm:"type:timestamp with time zone"`
	NextReviewDate *time.Time `json:"next_review_date" gorm:"type:timestamp with time zone"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}

// Track individual product purchases for product-based upgrade rules
type CustomerProductPurchaseStats struct {
	ID                        uuid.UUID `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	CustomerMembershipStatsID uuid.UUID `json:"customer_membership_stats_id" gorm:"type:uuid"`
	ProductID                 uuid.UUID `json:"product_id" gorm:"type:uuid;index"`
	Product                   *Product  `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	// Purchase counts for different review periods
	TotalQuantity      int `json:"total_quantity" gorm:"type:int;default:0"`
	MonthlyQuantity    int `json:"monthly_quantity" gorm:"type:int;default:0"`
	QuarterlyQuantity  int `json:"quarterly_quantity" gorm:"type:int;default:0"`
	YearlyQuantity     int `json:"yearly_quantity" gorm:"type:int;default:0"`
	HalfYearlyQuantity int `json:"half_yearly_quantity" gorm:"type:int;default:0"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}
