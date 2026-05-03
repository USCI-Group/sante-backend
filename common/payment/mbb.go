package payment

import (
	"context"
	"fmt"
	"encore.app/database/models"
)

type MaybankProvider struct {
	MerchantID string
	SecretKey  string
}

func (m *MaybankProvider) InitiatePayment(ctx context.Context, order *models.Order) (*InitiationResponse, error) {
	// TODO: Implement HMAC signature logic once MBB provided specs
	// reference: Order.OrderNumber
	// amount: Order.RoundedNetTotal
	
	txnID := fmt.Sprintf("MBB-%s-%s", order.OrderNumber, order.ID.String()[:8])
	
	return &InitiationResponse{
		PaymentURL:      "https://test.maybank2u.com.my/fpx/pay", // Temporary UAT URL
		TransactionID:   txnID,
		ReferenceNumber: order.OrderNumber,
		FormFields: map[string]string{
			"merchant_id": m.MerchantID,
			"amount":      fmt.Sprintf("%.2f", order.RoundedNetTotal),
			"order_no":    order.OrderNumber,
			"signature":   "PENDING_SPEC_LOGIC",
		},
	}, nil
}

func (m *MaybankProvider) VerifyWebhook(ctx context.Context, data map[string]string) (*WebhookResult, error) {
	// TODO: Verify signature from raw data
	return &WebhookResult{
		Success: true,
		Status:  "completed",
	}, nil
}

func (m *MaybankProvider) CheckStatus(ctx context.Context, ref string) (*StatusResult, error) {
	return &StatusResult{Status: "completed"}, nil
}
