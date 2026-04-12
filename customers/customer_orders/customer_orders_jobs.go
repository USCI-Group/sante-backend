package customer_orders

import (
	"log"
	"time"

	"encore.app/customers/customer_common"
)

var (
	SLEEP_TIME_CHECKING_QUEUE = 1 * time.Minute
)

func (s *Service) StartCustomerOrdersAutoCancelJobWorkers() error {
	// Start the worker in a goroutine without a transaction
	go func() {
		s.processCustomerOrdersAutoCancelJobFromQueue(SLEEP_TIME_CHECKING_QUEUE)
	}()
	log.Println("Customer orders auto cancel job workers started successfully")
	return nil
}

func (s *Service) processCustomerOrdersAutoCancelJobFromQueue(
	sleepTime time.Duration,
) {
	for {
		// Create a new transaction for each iteration
		trx := s.db.Begin()
		if trx.Error != nil {
			//log.Println("Error starting transaction:", trx.Error)
			time.Sleep(sleepTime)
			continue
		}

		timeLimit := 5 * time.Minute
		orderIDs, err := customer_common.GetOrderIDsThatNeedToBeAutoCancelled(trx, timeLimit)
		if err != nil {
			//log.Println("Error getting orders to auto cancel:", err)
			trx.Rollback()
			time.Sleep(sleepTime)
			continue
		}

		if len(orderIDs) == 0 {
			trx.Rollback()
			time.Sleep(sleepTime)
			continue // Changed from return nil to continue
		}

		err = customer_common.AutoCancelOrdersAndTransactions(trx, orderIDs)
		if err != nil {
			//log.Println("Error auto cancelling orders:", err)
			trx.Rollback()
			time.Sleep(sleepTime)
			continue
		}

		// Commit the transaction
		if err := trx.Commit().Error; err != nil {
			//log.Println("Error committing transaction:", err)
			trx.Rollback()
		} else {
			//log.Printf("Successfully auto cancelled %d orders", len(orderIDs))
		}

		time.Sleep(sleepTime)
	}
}
