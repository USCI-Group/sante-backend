package customer_payments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"encore.app/common/constants"
	"encore.app/common/payment"
	"encore.app/common_operations"
	"encore.app/customers/customer_common"
	"encore.app/database"
	"encore.app/database/models"
	"encore.dev/beta/auth"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

var secretsKeys struct {
	jwtSecretKey string
}

// initService initializes the user service.
// It is automatically called by Encore on service startup.
func initService() (*Service, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database.GetSanteDB().Stdlib(),
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	service := &Service{db: db}
	// Now call the method on the created service
	err = service.StartNotificationQueueJobWorkers()
	if err != nil {
		return nil, err
	}
	secretsKeys.jwtSecretKey = os.Getenv("JWT_SECRET_KEY")
	return &Service{db: db}, nil
}

type InitiatePaymentRequest struct {
	OrderID       uuid.UUID        `json:"order_id"`
	PaymentMethod string           `json:"payment_method"`
	Provider      payment.Provider `json:"provider"`
}

type InitiatePaymentResponse struct {
	PaymentMethod     constants.PaymentMethod `json:"payment_method"`
	ImageURL          string                  `json:"image_url"`
	PaymentURL        string                  `json:"payment_url"`
	TransactionID     string                  `json:"transaction_id"`
	TransactionNumber string                  `json:"transaction_number"`
	FormFields        map[string]string       `json:"form_fields,omitempty"`
}

type CheckPaymentStatusResponse struct {
	Message string                         `json:"message"`
	Data    CheckPaymentStatusResponseData `json:"data"`
	Success bool                           `json:"success"`
}

type CheckPaymentStatusResponseData struct {
	OrderID           uuid.UUID  `json:"order_id"`
	TransactionID     uuid.UUID  `json:"transaction_id"`
	TransactionNumber *string    `json:"transaction_number,omitempty"`
	PaymentAmount     float32    `json:"payment_amount"`
	PaymentMethod     string     `json:"payment_method"`
	PaymentStatus     string     `json:"payment_status"`
	OrderStatus       string     `json:"order_status"`
	TransactionDate   *time.Time `json:"transaction_date,omitempty"`
	LastChecked       time.Time  `json:"last_checked"`
	PointsRewarded    int        `json:"points_rewarded"`
}

type GetPaymentMethodsRequest struct {
	PaymentMethod    *models.PaymentMethod    `json:"payment_method"`
	PaymentChannel   *models.PaymentChannel   `json:"payment_channel"`
	PaymentPlatform  *models.PaymentPlatform  `json:"payment_platform"`
	IsActive         *bool                    `json:"is_active"`
	IsMaintenance    *bool                    `json:"is_maintenance"`
	IsVisible        *bool                    `json:"is_visible"`
	ValidFrom        *time.Time               `json:"valid_from"`
	ValidUntil       *time.Time               `json:"valid_until"`
	ComplianceStatus *models.ComplianceStatus `json:"compliance_status"`
}

type GetPaymentMethodsResponse struct {
	Message        string                              `json:"message"`
	PaymentMethods []models.PaymentMethodConfiguration `json:"payment_methods,omitempty"`
}

type RepaymentRequest struct {
	CustomerID uuid.UUID `json:"customer_id"`
	OrderID    uuid.UUID `json:"order_id"`
}

type RepaymentResponse struct {
	Message           string    `json:"message"`
	OrderID           uuid.UUID `json:"order_id"`
	PaymentLink       string    `json:"payment_link"`
	PaymentMethod     string    `json:"payment_method"`
	PaymentURL        string    `json:"payment_url"`
	TransactionID     string    `json:"transaction_id"`
	TransactionNumber string    `json:"transaction_number"`
}

func getGateway(provider payment.Provider) (payment.PaymentGateway, error) {
	switch provider {
	case payment.Maybank:
		return &payment.MaybankProvider{MerchantID: os.Getenv("MBB_MERCHANT_ID"), SecretKey: os.Getenv("MBB_SECRET_KEY")}, nil
	case payment.PublicBank:
		return &payment.PublicBankProvider{MerchantID: os.Getenv("PBB_MERCHANT_ID"), SecretKey: os.Getenv("PBB_SECRET_KEY")}, nil
	case payment.CyberSource:
		return &payment.CyberSourceProvider{
			ProfileID:   os.Getenv("CYBERSOURCE_PROFILE_ID"),
			AccessKey:   os.Getenv("CYBERSOURCE_ACCESS_KEY"),
			SecretKey:   os.Getenv("CYBERSOURCE_SECRET_KEY"),
			EndpointURL: os.Getenv("CYBERSOURCE_ENDPOINT_URL"),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// Internal function that can work with an existing transaction or create its own
func InitiatePaymentWithTx(ctx context.Context, req *InitiatePaymentRequest, existingTrx *gorm.DB) (*InitiatePaymentResponse, error) {
	var trx *gorm.DB

	if existingTrx != nil {
		trx = existingTrx
	} else {
		return nil, fmt.Errorf("transaction is required")
	}

	order, err := common_operations.GetOrderByID(trx, req.OrderID, false, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %v", err)
	}

	gateway, err := getGateway(req.Provider)
	if err != nil {
		return nil, err
	}

	initResponse, err := gateway.InitiatePayment(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate payment: %v", err)
	}

	// Create or update transaction record
	transaction := &models.Transaction{
		OrderID:              order.ID,
		GatewayTransactionID: &initResponse.TransactionID,
		TransactionNumber:    &initResponse.ReferenceNumber,
		PaymentURL:           &initResponse.PaymentURL,
		Amount:               order.RoundedNetTotal,
		PaymentMethod:        req.PaymentMethod,
		PaymentStatus:        models.PaymentStatusPending,
		TransactionDate:      time.Now(),
	}

	if err := trx.Create(transaction).Error; err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	return &InitiatePaymentResponse{
		PaymentMethod:     constants.PaymentMethod(req.PaymentMethod),
		ImageURL:          "",
		PaymentURL:        initResponse.PaymentURL,
		TransactionID:     initResponse.TransactionID,
		TransactionNumber: initResponse.ReferenceNumber,
		FormFields:        initResponse.FormFields,
	}, nil
}

// API to get payment link
//
//encore:api auth method=POST path=/api/customers/payments/initiate
func (s *Service) InitiatePayment(ctx context.Context, req *InitiatePaymentRequest) (*InitiatePaymentResponse, error) {
	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}
	defer trx.Rollback()

	resp, err := InitiatePaymentWithTx(ctx, req, trx)
	if err != nil {
		return nil, err
	}

	if err := trx.Commit().Error; err != nil {
		return nil, err
	}

	return resp, nil
}

// Check payment status
//
//encore:api auth method=GET path=/api/customers/payments/status/check/:order_id
func (s *Service) CheckPaymentStatus(ctx context.Context, order_id uuid.UUID) (*CheckPaymentStatusResponse, error) {
	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}
	defer trx.Rollback()

	transaction, err := customer_common.GetTransactionByOrderID(trx, order_id)
	if err != nil {
		trx.Rollback()
		return nil, err
	}
	// check if the payment is completed
	if transaction.PaymentStatus != "completed" {
		// Get payment status from PG
		// PG Inquiry removed
		/*
		resp := fiuu_payment.GetPaymentInquiryResponse{}
		if transaction.MolTransactionID != nil {
			temp_resp, err := fiuu_payment.GetPaymentInquiry(*transaction.MolTransactionID, transaction.Amount, constants.PaymentMethod(transaction.PaymentMethod), *merchantSecret)
			if err != nil {
				trx.Rollback()
				return nil, err
			}
			resp = *temp_resp
		} else {
			temp_resp, err := fiuu_payment.GetPaymentInquiryByTransactionNumber(*transaction.TransactionNumber, transaction.Amount, constants.PaymentMethod(transaction.PaymentMethod), *merchantSecret)
			if err != nil {
				trx.Rollback()
				return nil, err
			}
			resp = *temp_resp
		}
		*/

		// Update payment status based on PG response
		// check success status logic removed
		/*
		if resp.StatCode == fiuu_payment.FIUU_TRANSACTION_STATUS_SUCCESS {
			transaction.PaymentStatus = models.PaymentStatusCompleted
			transaction.Order.PaymentStatus = models.PaymentStatusCompleted
			transaction.Order.AmountReceived = &transaction.Order.RoundedNetTotal
		} else if resp.StatCode == fiuu_payment.FIUU_TRANSACTION_STATUS_FAILED {
			transaction.PaymentStatus = models.PaymentStatusFailed
			trx.Rollback()

			return &CheckPaymentStatusResponse{
				Message: "Payment not completed",
				Data: CheckPaymentStatusResponseData{
					OrderID:         order_id,
					TransactionID:   transaction.ID,
					PaymentAmount:   transaction.Amount,
					PaymentMethod:   transaction.PaymentMethod,
					PaymentStatus:   transaction.PaymentStatus,
					OrderStatus:     transaction.Order.OrderStatus,
					TransactionDate: &transaction.TransactionDate,
					LastChecked:     time.Now(),
				},
				Success: false,
			}, nil
		}
		*/
	}

	// Membership point reward logic removed
	pointsRewardedThisTransaction := 0
	if err := trx.Commit().Error; err != nil {
		trx.Rollback()
		return nil, err
	}

	customer_common.SendNotificationForPayment(
		ctx,
		s.db,
		transaction.OrderID,
		models.PaymentMethod(transaction.PaymentMethod),
		models.MembershipAppOrderNotification,
		transaction,
		&transaction.Order,
	)

	/* customer_common.SendNotificationForPayment(
		ctx,
		s.db,
		transaction.OrderID,
		models.PaymentMethod(transaction.PaymentMethod),
		models.MembershipAppOrderNotification,
		transaction,
		&transaction.Order,
	) */

	/* title := "Hooray There is a new order"
	body := "Order #" + transaction.Order.OrderNumber + " received. View details in the app."
	actionURL := "/streetfood/order_overview"
	// send notification to the POS
	err = pos.SendNotificationToPOS(
		ctx,
		&pos.SendNotificationToPOSRequest{
			Order:            &transaction.Order,
			Title:            &title,
			Body:             &body,
			ActionURL:        &actionURL,
			NotificationType: models.MembershipAppOrderNotification,
		},
	)
	firebase.SendOrderConfirmedNotification(ctx, s.db, transaction.OrderID) */

	return &CheckPaymentStatusResponse{
		Message: "Payment completed",
		Data: CheckPaymentStatusResponseData{
			OrderID:         order_id,
			TransactionID:   transaction.ID,
			PaymentAmount:   transaction.Amount,
			PaymentMethod:   transaction.PaymentMethod,
			PaymentStatus:   transaction.PaymentStatus,
			OrderStatus:     transaction.Order.OrderStatus,
			TransactionDate: &transaction.TransactionDate,
			LastChecked:     time.Now(),
			PointsRewarded:  pointsRewardedThisTransaction,
		},
		Success: true,
	}, nil

}

// API to get payment methods
//
//encore:api auth method=POST path=/api/customers/payments/methods/active/all
func (s *Service) GetPaymentMethods(ctx context.Context, req *GetPaymentMethodsRequest) (*GetPaymentMethodsResponse, error) {
	customerData, err := customer_common.GetCustomerDataFromAuthData(auth.Data)
	if err != nil {
		return nil, err
	}

	if req.IsActive == nil {
		req.IsActive = new(bool)
		*req.IsActive = true
	}
	if req.IsMaintenance == nil {
		req.IsMaintenance = new(bool)
		*req.IsMaintenance = false
	}
	if req.IsVisible == nil {
		req.IsVisible = new(bool)
		*req.IsVisible = true
	}
	fmt.Println("req.PaymentMethod", req.PaymentMethod)
	paymentMethods, err := customer_common.GetPaymentMethodsByBusinessID(
		s.db,
		*customerData.BusinessID,
		req.PaymentMethod,
		req.PaymentChannel,
		req.PaymentPlatform,
		*req.IsActive,
		*req.IsMaintenance,
		*req.IsVisible,
		req.ValidFrom,
		req.ValidUntil,
		req.ComplianceStatus,
	)
	if err != nil {
		return nil, err
	}
	return &GetPaymentMethodsResponse{
		Message:        "Payment methods fetched successfully",
		PaymentMethods: *paymentMethods,
	}, nil
}

// API to receive bank callback (Webhook)
//
//encore:api public raw method=POST path=/api/customers/payments/webhook/:provider
func (s *Service) PaymentWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	provider := ""
	// Path parameters in raw endpoints are available via r.PathValue in Go 1.22+ 
	// or Encore might provide them in a different way. 
	// For Encore raw endpoints, we usually use the URL.
	segments := strings.Split(r.URL.Path, "/")
	if len(segments) > 0 {
		provider = segments[len(segments)-1]
	}

	// Parse JSON body
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	trx := s.db.Begin()
	if trx.Error != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer trx.Rollback()

	gateway, err := getGateway(payment.Provider(provider))
	if err != nil {
		http.Error(w, "invalid provider", http.StatusBadRequest)
		return
	}

	result, err := gateway.VerifyWebhook(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if result.Success {
		var transaction models.Transaction
		if err := trx.Where("gateway_transaction_id = ?", result.TransactionID).First(&transaction).Error; err != nil {
			http.Error(w, "transaction not found", http.StatusNotFound)
			return
		}

		transaction.PaymentStatus = models.PaymentStatusCompleted
		if err := trx.Save(&transaction).Error; err != nil {
			http.Error(w, "failed to update transaction", http.StatusInternalServerError)
			return
		}

		// Update order status
		if err := trx.Model(&models.Order{}).Where("id = ?", transaction.OrderID).Update("status", models.OrderStatusPreparing).Error; err != nil {
			http.Error(w, "failed to update order", http.StatusInternalServerError)
			return
		}

		// Notify POS
		customer_common.SendNotificationForPayment(
			ctx,
			trx,
			transaction.OrderID,
			models.PaymentMethod(transaction.PaymentMethod),
			models.MembershipAppOrderNotification,
			&transaction,
			nil,
		)
	}

	if err := trx.Commit().Error; err != nil {
		http.Error(w, "commit error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// API to repayment for existing order
//
//encore:api auth method=POST path=/api/customers/payments/repayment
func (s *Service) Repayment(ctx context.Context, req *RepaymentRequest) (*RepaymentResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}
	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	// get order by order id
	order, err := common_operations.GetOrderByID(trx, req.OrderID, false, false)
	if err != nil {
		return nil, err
	}

	paymentMethod := InitiatePaymentRequest{
		OrderID:       order.ID,
		PaymentMethod: string(constants.PaymentMethodFPX),
		Provider:      payment.Maybank, // Default to Maybank for repayment placeholder
	}
	// get the payment link using the existing transaction
	paymentLink, err := InitiatePaymentWithTx(ctx, &paymentMethod, trx)
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	err = trx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &RepaymentResponse{
		Message:           "Repayment link created successfully",
		OrderID:           order.ID,
		PaymentLink:       paymentLink.PaymentURL,
		PaymentMethod:     string(paymentLink.PaymentMethod),
		PaymentURL:        paymentLink.PaymentURL,
		TransactionID:     paymentLink.TransactionID,
		TransactionNumber: paymentLink.TransactionNumber,
	}, nil
}
