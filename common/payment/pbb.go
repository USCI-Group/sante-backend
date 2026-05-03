package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"encore.app/database/models"
)

type PublicBankProvider struct {
	MerchantID string
	SecretKey  string
}

func (p *PublicBankProvider) InitiatePayment(ctx context.Context, order *models.Order) (*InitiationResponse, error) {
	txnID := fmt.Sprintf("PBB-%s-%s", order.OrderNumber, order.ID.String()[:8])
	amountString := fmt.Sprintf("%.2f", order.RoundedNetTotal)

	formFields := map[string]string{
		"seller_id": p.MerchantID,
		"txn_amt":   amountString,
		"txn_id":    txnID,
		"currency":  "MYR",
	}

	// Generate Signature
	signature := p.GenerateSignature(formFields)
	formFields["check_sum"] = signature

	return &InitiationResponse{
		PaymentURL:      "https://test.pbebank.com/fpx/pay", // Permanent UAT URL
		TransactionID:   txnID,
		ReferenceNumber: order.OrderNumber,
		FormFields:      formFields,
	}, nil
}

func (p *PublicBankProvider) VerifyWebhook(ctx context.Context, data map[string]string) (*WebhookResult, error) {

	receivedSignature := data["check_sum"]
	if receivedSignature == "" {
		return nil, fmt.Errorf("missing check_sum in webhook")
	}

	// Sign received data to verify
	expectedSignature := p.GenerateSignature(data)
	if !hmac.Equal([]byte(receivedSignature), []byte(expectedSignature)) {
		return &WebhookResult{Success: false, Status: "failed_signature"}, fmt.Errorf("invalid signature")
	}

	status := data["status"]
	success := status == "00" // Standard success code for many banks

	return &WebhookResult{
		Success:       success,
		TransactionID: data["txn_id"],
		Status:        status,
		RawResponse:   fmt.Sprintf("status=%s, reason=%s", status, data["reason"]),
	}, nil
}

func (p *PublicBankProvider) CheckStatus(ctx context.Context, ref string) (*StatusResult, error) {
	// PBB Inquiry API would go here
	return &StatusResult{Status: "pending-inquiry"}, nil
}

func (p *PublicBankProvider) GenerateSignature(fields map[string]string) string {
	// Exclude signature field itself from calculation
	var keys []string
	for k := range fields {
		if k != "check_sum" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fields[k])
	}
	
	// Append secret key at the end (standard HMAC pattern)
	data := sb.String()
	
	h := hmac.New(sha256.New, []byte(p.SecretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
