package models

import (
	"time"

	"encore.dev/types/uuid"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ActionType string

const (
	// Core Transaction Actions
	ActionTypePurchase ActionType = "purchase"
	ActionTypeRedeem   ActionType = "redeem"
	ActionTypeRefund   ActionType = "refund"

	// Enhanced Earning Actions
	ActionTypeFirstPurchase ActionType = "first_purchase"
	ActionTypeBirthdayBonus ActionType = "birthday_bonus"
	ActionTypeAnniversary   ActionType = "anniversary"
	ActionTypeReferral      ActionType = "referral"
	ActionTypeSocialMedia   ActionType = "social_media"
	ActionTypeReview        ActionType = "review"
	ActionTypeFeedback      ActionType = "feedback"
	ActionTypeSurvey        ActionType = "survey"
	ActionTypeMissionReward ActionType = "mission_reward"

	// Tier-Based Actions
	ActionTypeTierUpgrade     ActionType = "tier_upgrade"
	ActionTypeTierMaintenance ActionType = "tier_maintenance"
	ActionTypeTierBonus       ActionType = "tier_bonus"

	// Seasonal & Promotional Actions
	ActionTypeSeasonalBonus ActionType = "seasonal_bonus"
	ActionTypeFlashSale     ActionType = "flash_sale"
	ActionTypeWeekendBonus  ActionType = "weekend_bonus"
	ActionTypeHappyHour     ActionType = "happy_hour"

	// Engagement & Loyalty Actions
	ActionTypeDailyCheckin    ActionType = "daily_checkin"
	ActionTypeWeeklyVisit     ActionType = "weekly_visit"
	ActionTypeMonthlySpending ActionType = "monthly_spending"
	ActionTypeStreakBonus     ActionType = "streak_bonus"
	ActionTypeMilestone       ActionType = "milestone"

	// Product-Specific Actions
	ActionTypeNewProduct     ActionType = "new_product"
	ActionTypeCategoryBonus  ActionType = "category_bonus"
	ActionTypeBundlePurchase ActionType = "bundle_purchase"
	ActionTypeCrossSell      ActionType = "cross_sell"

	// Payment & Delivery Actions
	ActionTypeOnlineOrder   ActionType = "online_order"
	ActionTypeDeliveryOrder ActionType = "delivery_order"
	ActionTypePreOrder      ActionType = "pre_order"
	ActionTypeSubscription  ActionType = "subscription"
	ActionTypeAutoRenewal   ActionType = "auto_renewal"

	// Customer Service Actions
	ActionTypeComplaint   ActionType = "complaint"
	ActionTypeSuggestion  ActionType = "suggestion"
	ActionTypeTestimonial ActionType = "testimonial"
	ActionTypePhotoShare  ActionType = "photo_share"
)

type PointRule struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	MembershipID       *uuid.UUID     `json:"membership_id" gorm:"type:uuid"`
	Membership         *Membership    `json:"membership" gorm:"foreignKey:MembershipID"`
	BusinessID         *uuid.UUID     `json:"business_id" gorm:"type:uuid"`
	Name               string         `json:"name" gorm:"type:varchar(255)"`
	ActionType         ActionType     `json:"action_type" gorm:"type:varchar(255)"`
	PointsEarned       *int           `json:"points_earned" gorm:"type:int" validate:"gte=0"`
	PointsMultiplier   *int           `json:"points_multiplier" gorm:"type:int" validate:"gte=0"`
	MinAmount          *float32       `json:"min_amount" gorm:"type:numeric" validate:"gte=0"`
	MaxPointsPerAction *int           `json:"max_points_per_action" gorm:"type:int"`
	MaxPointsPerDay    *int           `json:"max_points_per_day" gorm:"type:int"`
	MaxPointsPerMonth  *int           `json:"max_points_per_month" gorm:"type:int"`
	MaxPointsPerYear   *int           `json:"max_points_per_year" gorm:"type:int"`
	DayOfWeek          *int           `json:"day_of_week" gorm:"type:int"` // This is the day ordering, it can be more than 7 days.
	IsActive           bool           `json:"is_active" gorm:"type:boolean"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          *time.Time     `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at,omitempty"`
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func (p *PointRule) Validate() error {
	return validate.Struct(p)
}
