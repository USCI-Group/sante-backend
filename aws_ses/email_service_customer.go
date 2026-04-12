package aws_ses

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"encore.app/common"
	"encore.app/database/models"
)

func SendDeactivateAccountEmail(customer models.Customer, companyName string) error {

	recipientName := customer.FirstName + " " + customer.LastName
	reactivateLink, err := GenerateReactivateLink(customer)
	if err != nil {
		return err
	}

	htmlBody := strings.Replace(DeactivateAccountTemplate, "{{name}}", recipientName, 1)
	htmlBody = strings.Replace(htmlBody, "{{reactivate_link}}", *reactivateLink, 1)
	htmlBody = strings.Replace(htmlBody, "{{year}}", strconv.Itoa(time.Now().Year()), 1)
	htmlBody = strings.ReplaceAll(htmlBody, "{{company_name}}", companyName)

	err = SendRawEmail(customer.Email, "Account Deactivation Notice", htmlBody)
	if err != nil {
		return err
	}

	return nil
}

func GenerateReactivateLink(customer models.Customer) (*string, error) {
	token, err := common.GenerateJWTToken(common.TokenOptions{
		UserID:              customer.ID,
		ExpiryTimeInMinutes: 60 * 24 * 30, // 30 days
	})
	if err != nil {
		return nil, err
	}

	link := fmt.Sprintf("%s/api/account/reactivate/%s", common.GetOrigin(), *token)

	return &link, nil
}

func SendReactivateAccountEmail(customer models.Customer, companyName string) error {
	recipientName := customer.FirstName + " " + customer.LastName

	htmlBody := strings.Replace(ReactivateAccountTemplate, "{{name}}", recipientName, 1)
	htmlBody = strings.Replace(htmlBody, "{{year}}", strconv.Itoa(time.Now().Year()), 1)
	htmlBody = strings.ReplaceAll(htmlBody, "{{company_name}}", companyName)

	err := SendRawEmail(customer.Email, "Account Reactivation", htmlBody)
	if err != nil {
		return err
	}

	return nil
}

func SendEInvoiceGeneratedEmail(email string, name string, invoiceNumber string, invoiceUrl string) error {
	htmlBody := strings.Replace(EInvoiceGeneratedTemplate, "{{name}}", name, 1)
	htmlBody = strings.Replace(htmlBody, "{{invoice_number}}", invoiceNumber, 1)
	htmlBody = strings.Replace(htmlBody, "{{invoice_url}}", invoiceUrl, 1)

	err := SendRawEmail(email, "E-Invoice Generated", htmlBody)
	if err != nil {
		log.Println("[aws_ses.SendEInvoiceGeneratedEmail] Error: Failed to send E-Invoice generated email:", err)
		return err
	}

	return nil
}
