package customer_payments

import (
	"context"
	"log"
	"time"

	"encore.app/customers/customer_common"
)

var (
	SLEEP_TIME_CHECKING_QUEUE = 1 * time.Minute
)

func (s *Service) StartNotificationQueueJobWorkers() error {
	// Start the worker in a goroutine without a transaction
	go func() {
		s.processNotificationQueueJobFromQueue(SLEEP_TIME_CHECKING_QUEUE)
	}()
	log.Println("Notification queue job workers started successfully")
	return nil
}

func (s *Service) processNotificationQueueJobFromQueue(
	sleepTime time.Duration,
) {
	for {
		ctx := context.Background()
		notificationQueues, err := customer_common.GetNotificationQueue(s.db)
		if len(notificationQueues) == 0 {
			time.Sleep(sleepTime)
			continue
		}
		if err != nil {
			time.Sleep(sleepTime)
			continue
		}

		for _, notificationQueue := range notificationQueues {
			trx := s.db.Begin()
			if trx.Error != nil {
				trx.Rollback()
				continue
			}
			order, err := customer_common.GetOrderByID(trx, notificationQueue.OrderID, true)
			if err != nil {
				trx.Rollback()
				continue
			}
			// send notification to the POS
			err = customer_common.SendNotificationToPOS(
				ctx,
				trx,
				&customer_common.SendNotificationToPOSRequest{
					Order:            order,
					Title:            &notificationQueue.Title,
					Body:             &notificationQueue.Body,
					ActionURL:        &notificationQueue.ActionURL,
					NotificationType: notificationQueue.NotificationType,
				},
			)
			if err != nil {
				trx.Rollback()
				continue
			}
			result := customer_common.DeleteNotificationQueue(trx, notificationQueue.ID, true)
			if result != nil {
				trx.Rollback()
				continue
			}
			// Commit the transaction
			commitErr := trx.Commit().Error
			if commitErr != nil {
				trx.Rollback()
				continue
			}
		}
		time.Sleep(sleepTime)
	}
}
