package customer_common

import (
	"time"

	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

func AddNotificationToQueue(
	trx *gorm.DB,
	order_id uuid.UUID,
	notification_type models.NotificationType,
	title string,
	body string,
	actionURL string,
	planToSendAt time.Time,
) {
	notificationQueue := &models.NotificationQueue{
		OrderID:          order_id,
		QueueType:        models.QueueTypePickupLater,
		ImageURL:         nil,
		NotificationType: notification_type,
		Title:            title,
		Body:             body,
		ActionURL:        actionURL,
		PlanToSendAt:     planToSendAt,
		QueueStatus:      models.QueueStatusPending,
		SendAt:           nil,
		CompletedAt:      nil,
		FailedAt:         nil,
		CancelledAt:      nil,
		CreatedAt:        time.Now(),
	}
	trx.Create(notificationQueue)
}

func GetNotificationQueue(
	trx *gorm.DB,
) ([]models.NotificationQueue, error) {
	timeNow := time.Now().UTC()
	var notificationQueues []models.NotificationQueue
	result := trx.Where("plan_to_send_at <= ?", timeNow).Find(&notificationQueues)
	if result.Error != nil {
		return nil, result.Error
	}
	return notificationQueues, nil
}

func DeleteNotificationQueue(
	trx *gorm.DB,
	notification_queue_id uuid.UUID,
	isHardDelete bool,
) error {
	if isHardDelete {
		result := trx.Where("id = ?", notification_queue_id).Unscoped().Delete(&models.NotificationQueue{})
		if result.Error != nil {
			return result.Error
		}
	} else {
		result := trx.Where("id = ?", notification_queue_id).Delete(&models.NotificationQueue{})
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}
