package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type NotificationType string

const (
	PromotionNotification              NotificationType = "promotion"
	OrderNotification                  NotificationType = "order"
	OutletNotification                 NotificationType = "outlet"
	UserNotification                   NotificationType = "user"
	GeneralNotification                NotificationType = "general"
	AlertNotification                  NotificationType = "alert"
	CriticalNotification               NotificationType = "critical"
	ChatNotification                   NotificationType = "chat"
	PaymentNotification                NotificationType = "payment"
	DeliveryNotification               NotificationType = "delivery"
	PickupNotification                 NotificationType = "pickup"
	OtherNotification                  NotificationType = "other"
	GrabFoodNotification               NotificationType = "grabfood"
	GrabFoodOrderCancelledNotification NotificationType = "grabfoodordercancelled"
	ShopeeFoodNotification             NotificationType = "shopeefood"
	StockRequestNotification           NotificationType = "stockrequest"
	MembershipAppOrderNotification     NotificationType = "membershipapporder"
	CampaignPushNotificationType       NotificationType = "campaignpushnotification"
)

type Notification struct {
	ID               uuid.UUID        `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	UserID           *uuid.UUID       `json:"user_id" gorm:"type:uuid"`
	User             *User            `json:"user" gorm:"foreignKey:UserID"`
	OutletID         *uuid.UUID       `json:"outlet_id" gorm:"type:uuid"`
	Outlet           *Outlet          `json:"outlet" gorm:"foreignKey:OutletID"`
	FCMDeviceToken   *string          `json:"fcm_device_token" gorm:"type:varchar(255)"`
	Title            *string          `json:"title" gorm:"type:varchar(255)"`
	Body             *string          `json:"body" gorm:"type:text"`
	Data             *string          `json:"data" gorm:"type:jsonb"`
	NotificationType NotificationType `json:"notification_type" gorm:"type:text"`
	IsRead           bool             `json:"is_read" gorm:"type:boolean; default:false"`
	ActionURL        *string          `json:"action_url" gorm:"type:text"`
	ImageURL         *string          `json:"image_url" gorm:"type:text"`
	ExpiredAt        *time.Time       `json:"expired_at"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        *time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `json:"deleted_at,omitempty"`
}
