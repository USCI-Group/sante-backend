package models

import (
	"time"

	"encore.app/common"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type EmailActionType string

const (
	EmailActionTypeForgetPassword EmailActionType = "forget_password"
	EmailActionTypeVerifyEmail    EmailActionType = "verify_email"
)

type Customer struct {
	ID                        uuid.UUID                  `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID                uuid.UUID                  `json:"business_id" gorm:"type:uuid"`
	FirstName                 string                     `json:"first_name" gorm:"type:varchar(255)" valid:"required~FirstName is required"`
	LastName                  string                     `json:"last_name" gorm:"type:varchar(255)" valid:"required~LastName is required"`
	Email                     string                     `json:"email" gorm:"type:varchar(255)" valid:"required~Email is required"`
	Password                  string                     `json:"-" gorm:"type:varchar(255)" valid:"required~Password is required"`
	PhoneNumber               string                     `json:"phone_number" gorm:"type:varchar(255)" valid:"required~PhoneNumber is required"`
	ProfilePicture            string                     `json:"profile_picture" gorm:"type:varchar(500)"`
	DateOfBirth               time.Time                  `json:"date_of_birth" gorm:"type:date"`
	EmailVerified             bool                       `json:"email_verified" gorm:"type:boolean; default:false"`
	IsNewsletterSubscribed    bool                       `json:"is_newsletter_subscribed" gorm:"type:boolean; default:false"`
	IsAgreeToTerms            bool                       `json:"is_agree_to_terms" gorm:"type:boolean; default:false"`
	IsAgreeToPrivacyPolicy    bool                       `json:"is_agree_to_privacy_policy" gorm:"type:boolean; default:false"`
	CustomerFavouriteProducts []CustomerFavouriteProduct `json:"customer_favourite_products" gorm:"foreignKey:CustomerID"`
	CustomerDeliveryAddresses []CustomerDeliveryAddress  `json:"customer_delivery_addresses" gorm:"foreignKey:CustomerID"`
	IsActive                  bool                       `json:"is_active" gorm:"type:boolean; default:true"`
	CreatedAt                 time.Time                  `json:"created_at"`
	UpdatedAt                 *time.Time                 `json:"updated_at"`
	DeletedAt                 gorm.DeletedAt             `json:"deleted_at,omitempty"`
}

type CustomerDeliveryAddress struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	CustomerID uuid.UUID      `json:"customer_id" gorm:"type:uuid;index"`
	Address    common.Address `json:"address" gorm:"embedded"`
	Name       string         `json:"name" gorm:"type:varchar(255)"`
	IsDefault  bool           `json:"is_default" gorm:"type:boolean; default:false"`
	Latitude   *float64       `json:"latitude,omitempty" gorm:"type:decimal(12,8)"`
	Longitude  *float64       `json:"longitude,omitempty" gorm:"type:decimal(12,8)"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  *time.Time     `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type CustomerToken struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	CustomerID   uuid.UUID      `json:"customer_id" gorm:"type:uuid"`
	RefreshToken string         `json:"refresh_token" gorm:"type:varchar(255)"`
	ExpiredAt    time.Time      `json:"expired_at" gorm:"type:timestamp with time zone"`
	DeviceID     *string        `json:"device_id" gorm:"type:varchar(255)"`
	FCMToken     *string        `json:"fcm_token" gorm:"type:varchar(255)"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    *time.Time     `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type CustomerVoucher struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	CustomerID  uuid.UUID      `json:"customer_id" gorm:"type:uuid"`
	VoucherID   uuid.UUID      `json:"voucher_id" gorm:"type:uuid"`
	Voucher     Voucher        `json:"voucher" gorm:"foreignKey:VoucherID"`
	VoucherCode *string        `json:"voucher_code" gorm:"type:varchar(255)"`
	IsRedeemed  bool           `json:"is_redeemed" gorm:"type:boolean; default:true"`
	RedeemedAt  *time.Time     `json:"redeemed_at" gorm:"type:timestamp with time zone"`
	Used        bool           `json:"used" gorm:"type:boolean; default:false"`
	UsedAt      *time.Time     `json:"used_at" gorm:"type:timestamp with time zone"`
	Validity    *int           `json:"validity" gorm:"type:int"`
	ValidFrom   time.Time      `json:"valid_from" gorm:"type:timestamp with time zone"`
	ValidTo     time.Time      `json:"valid_to" gorm:"type:timestamp with time zone"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type CustomerMissionStatus string

const (
	CustomerMissionStatusInProgress CustomerMissionStatus = "in_progress"
	CustomerMissionStatusCompleted  CustomerMissionStatus = "completed"
	CustomerMissionStatusExpired    CustomerMissionStatus = "expired"
	CustomerMissionStatusCancelled  CustomerMissionStatus = "cancelled"
)

// One row per enrollment/attempt. Allows re-joining missions in future by creating a new attempt.
type CustomerMission struct {
	ID          uuid.UUID             `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID  uuid.UUID             `json:"business_id" gorm:"type:uuid"`
	CustomerID  uuid.UUID             `json:"customer_id" gorm:"type:uuid;index:idx_attempt_unique,unique"`
	MissionID   uuid.UUID             `json:"mission_id" gorm:"type:uuid;index:idx_attempt_unique,unique"`
	Mission     Mission               `json:"mission" gorm:"foreignKey:MissionID"`
	Status      CustomerMissionStatus `json:"status" gorm:"type:varchar(20);index"`
	Progress    int                   `json:"progress" gorm:"type:int;default:0"` // 0..100 snapshot (optional convenience)
	StartedAt   time.Time             `json:"started_at" gorm:"type:timestamp with time zone"`
	ExpiresAt   *time.Time            `json:"expires_at" gorm:"type:timestamp with time zone"`
	CompletedAt *time.Time            `json:"completed_at" gorm:"type:timestamp with time zone"`
	EndedAt     *time.Time            `json:"ended_at" gorm:"type:timestamp with time zone"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   *time.Time            `json:"updated_at"`
	DeletedAt   gorm.DeletedAt        `json:"deleted_at,omitempty"`
	// Prevent two concurrent in-progress attempts for same (customer, mission)
	// idx_attempt_unique should include (customer_id, mission_id, status) with a partial index on status='in_progress' at DB-level.
}

// Per criteria progress (one row per MissionCriteria)
type CustomerMissionCriteriaProgress struct {
	ID                uuid.UUID           `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	CustomerMissionID uuid.UUID           `json:"customer_mission_id" gorm:"type:uuid;index"`
	MissionCriteriaID uuid.UUID           `json:"mission_criteria_id" gorm:"type:uuid;index"`
	MissionCriteria   MissionCriteria     `json:"mission_criteria" gorm:"foreignKey:MissionCriteriaID"`
	CriteriaType      MissionCriteriaType `json:"criteria_type" gorm:"type:varchar(50)"`
	// For numeric/field-based criteria (order_count, spending_amount)
	CurrentNumeric float64 `json:"current_numeric" gorm:"type:decimal(10,2);default:0"`
	TargetNumeric  float64 `json:"target_numeric" gorm:"type:decimal(10,2);default:0"`
	// Snapshotted satisfaction
	IsSatisfied bool           `json:"is_satisfied" gorm:"type:boolean;default:false"`
	LastEventAt *time.Time     `json:"last_event_at" gorm:"type:timestamp with time zone"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}

// Idempotency for granting rewards once per attempt
type MissionRewardGrant struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	CustomerMissionID uuid.UUID      `json:"customer_mission_id" gorm:"type:uuid;index"`
	MissionRewardID   uuid.UUID      `json:"mission_reward_id" gorm:"type:uuid;index"`
	GrantedAt         time.Time      `json:"granted_at" gorm:"type:timestamp with time zone"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         *time.Time     `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at,omitempty"`
}
