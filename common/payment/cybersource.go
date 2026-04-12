package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"encore.app/database/models"
	"encore.dev/types/uuid"
)

type CyberSourceProvider struct {
	ProfileID   string
	AccessKey   string
	SecretKey   string
	EndpointURL string
}

func (p *CyberSourceProvider) InitiatePayment(ctx context.Context, order *models.Order) (*InitiationResponse, error) {
	// CyberSource uses UUIDs natively, so we'll generate one for the transaction reference
	txnUUID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate transaction UUID: %v", err)
	}

	signedDateTime := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	referenceNumber := order.OrderNumber

	// Basic required fields for Secure Acceptance Hosted Checkout
	formFields := map[string]string{
		"access_key":           p.AccessKey,
		"profile_id":           p.ProfileID,
		"transaction_uuid":     txnUUID.String(),
		"signed_field_names":   "access_key,profile_id,transaction_uuid,signed_field_names,unsigned_field_names,signed_date_time,locale,transaction_type,reference_number,amount,currency",
		"unsigned_field_names": "",
		"signed_date_time":     signedDateTime,
		"locale":               "en",
		"transaction_type":     "sale",
		"reference_number":     referenceNumber,
		"amount":               fmt.Sprintf("%.2f", order.RoundedNetTotal),
		"currency":             "MYR",
	}

	// Generate Signature
	signature := p.GenerateSignature(formFields)
	formFields["signature"] = signature

	// Using the provided endpoint from context, or default to CyberSource test environment if empty
	endpoint := p.EndpointURL
	if endpoint == "" {
		endpoint = "https://testsecureacceptance.cybersource.com/pay"
	}

	return &InitiationResponse{
		PaymentURL:      endpoint,
		TransactionID:   txnUUID.String(), // Using their UUID as gateway ID
		ReferenceNumber: referenceNumber,
		FormFields:      formFields,
	}, nil
}

func (p *CyberSourceProvider) VerifyWebhook(ctx context.Context, rawData interface{}) (*WebhookResult, error) {
	// Secure Acceptance Webhook sends data essentially as POST form variables,
	// Assuming rawData is a map[string]string constructed by the handler
	data, ok := rawData.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("invalid webhook data type, expected map[string]string")
	}

	// Extract the signature CyberSource sent us
	receivedSignature := data["signature"]
	if receivedSignature == "" {
		return nil, fmt.Errorf("missing signature in webhook payload")
	}

	// According to docs, they return 'signed_field_names' detailing what to sign to verify
	// Reconstruct the signature using only the fields they stated they signed
	reconstructedSignature := p.GenerateSignature(data)

	if receivedSignature != reconstructedSignature {
		return &WebhookResult{Success: false, Status: "failed_signature_validation"}, fmt.Errorf("webhook signature validation failed")
	}

	// Decision field defines the outcome (ACCEPT, REJECT, ERROR)
	decision := data["decision"]
	success := decision == "ACCEPT"
	
	status := "completed"
	if !success {
		status = "failed"
	}

	amountVal, _ := fmt.Sscanf(data["auth_amount"], "%f")

	return &WebhookResult{
		Success:       success,
		TransactionID: data["req_transaction_uuid"],
		Amount:        float64(amountVal),
		Status:        status,
		RawResponse:   fmt.Sprintf("decision=%s, reason_code=%s", decision, data["reason_code"]),
	}, nil
}

func (p *CyberSourceProvider) CheckStatus(ctx context.Context, transactionReference string) (*StatusResult, error) {
	// CyberSource Secure Acceptance Hosted Checkout primarily relies on the Webhook (Server-to-Server)
	// or returning Customer redirect for status, API polling requires different credentials (REST API).
	// Implementing a stub here as per standard.
	return &StatusResult{Status: "pending-webhook", Raw: ""}, nil
}

// GenerateSignature concatenates values as per 'signed_field_names' and hashes via HMAC-SHA256
func (p *CyberSourceProvider) GenerateSignature(fields map[string]string) string {
	signedFieldNamesStr, ok := fields["signed_field_names"]
	if !ok {
		return ""
	}

	fieldNames := strings.Split(signedFieldNamesStr, ",")
	var dataToSign []string

	for _, fieldName := range fieldNames {
		value := fields[fieldName]
		dataToSign = append(dataToSign, fmt.Sprintf("%s=%s", fieldName, value))
	}

	commaSeparatedData := strings.Join(dataToSign, ",")
	return p.signData(commaSeparatedData, p.SecretKey)
}

func (p *CyberSourceProvider) signData(data string, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
