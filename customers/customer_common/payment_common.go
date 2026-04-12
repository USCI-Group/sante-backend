package customer_common

import (
	"context"
	"fmt"
	"log"
	"time"

	"encore.app/common_operations"
	"encore.app/database/models"
	"encore.app/firebase"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type SendNotificationToPOSRequest struct {
	Order            *models.Order           `json:"order"`
	Title            *string                 `json:"title,omitempty"`
	Body             *string                 `json:"body,omitempty"`
	ActionURL        *string                 `json:"action_url,omitempty"`
	NotificationType models.NotificationType `json:"notification_type"`
}

func SendNotificationForPayment(
	ctx context.Context,
	trx *gorm.DB,
	order_id uuid.UUID,
	payment_method models.PaymentMethod,
	notification_type models.NotificationType,
	transaction *models.Transaction, // not necessarily to be provided
	order *models.Order, // not necessarily to be provided
) {
	if transaction == nil {
		// get transaction from database
		transaction = &models.Transaction{}
		if err := trx.Where("order_id = ?", order_id).First(transaction).Error; err != nil {
			fmt.Printf("[SendNotificationForPayment] error getting transaction: %v", err)
			return
		}
	}

	if order == nil {
		// get order from database
		order = &models.Order{}
		if err := trx.Where("id = ?", order_id).First(order).Error; err != nil {
			fmt.Printf("[SendNotificationForPayment] error getting order: %v", err)
			return
		}
	}
	title := "Hooray There is a new order"
	body := "Order #" + transaction.Order.OrderNumber + " received. View details in the app."
	actionURL := "/streetfood/order_overview"

	log.Printf("[SendNotificationForPayment] title: %v, body: %v, actionURL: %v", title, body, actionURL)
	timeNow := time.Now().UTC()
	posRequest := &SendNotificationToPOSRequest{
		Order:            order,
		Title:            &title,
		Body:             &body,
		ActionURL:        &actionURL,
		NotificationType: notification_type,
	}

	// When PickupAt is nil (ASAP order), send to POS only
	if order.PickupAt == nil {
		_ = SendNotificationToPOS(ctx, trx, posRequest)
	} else {
		orderPickupAt := order.PickupAt.UTC()
		timeDiff := orderPickupAt.Sub(timeNow)
		log.Printf("[SendNotificationForPayment] orderPickupAt: %v, timeNow: %v, timeDiff: %v", orderPickupAt, timeNow, timeDiff)

		if timeDiff > 0 {
			planToSendAt := orderPickupAt.Add(-30 * time.Minute)
			AddNotificationToQueue(trx, order_id, notification_type, title, body, actionURL, planToSendAt)
		} else {
			_ = SendNotificationToPOS(ctx, trx, posRequest)
		}
	}

	/// send notification to membership app removed
}

// SendNotificationToPOS sends notification to POS users about a new order
// This is a private API that can only be called from other services
func SendNotificationToPOS(
	ctx context.Context,
	trx *gorm.DB,
	req *SendNotificationToPOSRequest,
) error {
	order := req.Order
	title := req.Title
	body := req.Body
	actionURL := req.ActionURL
	notificationType := req.NotificationType

	// send notification to the customer
	// plan to do such as query device token from the database
	var merchantSecret models.MerchantSecret
	result := trx.Model(&models.MerchantSecret{}).Where("outlet_id = ?", order.OutletID).First(&merchantSecret)
	if result.Error != nil {
		return nil
	}

	var users []models.User
	result = trx.Model(&models.User{}).Where("outlet_id = ?", order.OutletID).Find(&users)
	if result.Error != nil {
		return nil
	}
	// insert notification into database (for all user which under a same outlet)
	var notificationIDs []uuid.UUID
	var deviceTokens []string

	for _, user := range users {
		notification, _ := common_operations.InsertNotification(trx, ctx, &models.Notification{
			OutletID:         &order.OutletID,
			UserID:           &user.ID,
			FCMDeviceToken:   user.FCMDeviceToken,
			Title:            title,
			Body:             body,
			NotificationType: models.MembershipAppOrderNotification,
			IsRead:           false,
			ActionURL:        nil,
			ImageURL:         nil,
			ExpiredAt:        nil,
		})
		notificationIDs = append(notificationIDs, notification.ID)
		if user.FCMDeviceToken != nil {
			deviceTokens = append(deviceTokens, *user.FCMDeviceToken)
		} else {
			deviceTokens = append(deviceTokens, "")
		}
	}

	firebase.SendNotificationToMultipleDevices(
		ctx,
		deviceTokens,
		*title,
		*body,
		notificationIDs,
		&req.Order.ID,
		actionURL,
		notificationType,
		firebase.FirebaseAppTypePOS,
		nil,
	)
	return nil
}
