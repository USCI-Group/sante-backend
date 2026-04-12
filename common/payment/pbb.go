package payment

import (
	"context"
	"fmt"
	"encore.app/database/models"
)

type PublicBankProvider struct {
	MerchantID string
	SecretKey  string
}

func (p *PublicBankProvider) InitiatePayment(ctx context.Context, order *models.Order) (*InitiationResponse, error) {
	txnID := fmt.Sprintf("PBB-%s-%s", order.OrderNumber, order.ID.String()[:8])
	
	return &InitiationResponse{
		PaymentURL:      "https://test.pbebank.com/fpx/pay", // Temporary UAT URL
		TransactionID:   txnID,
		ReferenceNumber: order.OrderNumber,
		FormFields: map[string]string{
			"seller_id": p.MerchantID,
			"txn_amt":   fmt.Sprintf("%.2f", order.RoundedNetTotal),
			"txn_id":    txnID,
		},
	}, nil
}

func (p *PublicBankProvider) VerifyWebhook(ctx context.Context, rawData interface{}) (*WebhookResult, error) {
	return &WebhookResult{
		Success: true,
		Status:  "completed",
	}, nil
}

func (p *PublicBankProvider) CheckStatus(ctx context.Context, ref string) (*StatusResult, error) {
	return &StatusResult{Status: "completed"}, nil
}
