package payment

import (
	"context"
	"encore.app/database/models"
)

// PaymentGateway defines the interface for all payment providers (Maybank, Public Bank, etc.)
type PaymentGateway interface {
	// InitiatePayment generates the redirect URL and any required form data
	InitiatePayment(ctx context.Context, order *models.Order) (*InitiationResponse, error)
	// VerifyWebhook processes the bank's callback data
	VerifyWebhook(ctx context.Context, rawData interface{}) (*WebhookResult, error)
	// CheckStatus manually queries the bank for transaction status
	CheckStatus(ctx context.Context, transactionReference string) (*StatusResult, error)
}

type InitiationResponse struct {
	PaymentURL      string            `json:"payment_url"`
	FormFields      map[string]string `json:"form_fields,omitempty"`
	TransactionID   string            `json:"transaction_id"`
	ReferenceNumber string            `json:"reference_number"`
}

type WebhookResult struct {
	Success          bool   `json:"success"`
	TransactionID    string `json:"transaction_id"`
	Amount           float64 `json:"amount"`
	Status           string  `json:"status"`
	RawResponse      string `json:"raw_response"`
}

type StatusResult struct {
	Status string `json:"status"`
	Raw    string `json:"raw"`
}

type Provider string

const (
	Maybank    Provider = "mbb"
	PublicBank Provider = "pbb"
	CyberSource Provider = "cybersource"
)
