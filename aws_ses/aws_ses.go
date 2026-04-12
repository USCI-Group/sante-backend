package aws_ses

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"encore.app/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

//encore:service
type Service struct {
	client *ses.SES
}

var SESClient *ses.SES

var awsConfig struct {
	Region          string `json:"aws_ses_region"`
	AwsSESKeyId     string `json:"aws_ses_key_id"`
	AwsSESSecretKey string `json:"aws_ses_secret_key"`
	AwsSessionToken string `json:"aws_session_token"`
	Sender          string `json:"sender"`
}

// TEST CONSTANT
const (
	Sender = "ichiroadris@gmail.com"
	//Sender = "mrchurros.dev@gmail.com"
	Recipient = "recipient@example.com"
	Subject   = "Amazon SES Test (AWS SDK for Go)"
	HtmlBody  = "<h1>Amazon SES Test Email (AWS SDK for Go)</h1><p>This email was sent with " +
		"<a href='https://aws.amazon.com/ses/'>Amazon SES</a> using the " +
		"<a href='https://aws.amazon.com/sdk-for-go/'>AWS SDK for Go</a>.</p>"
	TextBody = "This email was sent with Amazon SES using the AWS SDK for Go."
	CharSet  = "UTF-8"
)

// initService initializes the site service.
// It is automatically called by Encore on service startup.
func initService() (*Service, error) {
	common.LoadEnv()

	awsConfig.Region = os.Getenv("AWS_REGION")
	awsConfig.AwsSESKeyId = os.Getenv("AWS_SES_ACCESS_KEY_ID")
	awsConfig.AwsSESSecretKey = os.Getenv("AWS_SES_ACCESS_KEY")
	awsConfig.Sender = os.Getenv("AWS_SES_SENDER_EMAIL")
	//awsConfig.AwsSessionToken = os.Getenv("AWS_SES_SESSION_TOKEN")

	// Don't log the full secret key for security, just the first few characters
	if len(awsConfig.AwsSESSecretKey) > 4 {
		log.Printf("Secret Access Key: %s...", awsConfig.AwsSESSecretKey[:4])
	} else {
		log.Printf("Secret Access Key: [EMPTY OR TOO SHORT]")
	}

	// Add validation
	if awsConfig.Region == "" {
		fmt.Println("AWS_REGION environment variable is not set")
	}
	if awsConfig.AwsSESKeyId == "" {
		fmt.Println("AWS_SES_ACCESS_KEY_ID environment variable is not set")
	}
	if awsConfig.AwsSESSecretKey == "" {
		fmt.Println("AWS_SES_ACCESS_KEY environment variable is not set")
	}
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsConfig.Region),
		Credentials: credentials.NewStaticCredentials(awsConfig.AwsSESKeyId, awsConfig.AwsSESSecretKey, ""),
	})
	if err != nil {
		log.Println("AWS SES configuration error! Please check your configuration.")
		return nil, err
	}

	client := ses.New(session)

	// Initialize the global SESClient
	SESClient = client

	return &Service{client: client}, nil
}

const (
	EmailVerificationTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Please verify your email</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 0; }
    .container { max-width: 600px; background-color: #ffffff; margin: 40px auto; padding: 20px; border-radius: 8px; }
    .header { font-size: 24px; font-weight: bold; margin-bottom: 20px; }
    .content { font-size: 16px; line-height: 1.5; margin-bottom: 30px; }
    .verification-code { font-size: 32px; font-weight: bold; text-align: center; background-color: #f8f9fa; padding: 20px; border-radius: 8px; border: 2px dashed #dee2e6; margin: 20px 0; letter-spacing: 8px; color: #1a73e8; }
    .footer { font-size: 12px; color: #888888; margin-top: 30px; }
  </style>
</head>
<body>
  <table width="100%" cellpadding="0" cellspacing="0" role="presentation">
    <tr>
      <td align="center">
        <div class="container">
          <div class="header">Verify Your Email Address</div>
          <div class="content">
            Hello {{name}},<br><br>
            Thanks for signing up! To get started, please confirm your email address by entering the verification code below:
          </div>
          <div class="verification-code">
            {{verification_code}}
          </div>
          <div class="content">
            Enter this 6-digit code in the verification field to complete your registration.<br><br>
            This code will expire in 10 minutes for security reasons.
          </div>
          <div class="footer">
            &copy; {{year}} {{company_name}}. All rights reserved.<br>
            If you didn't sign up for this, just ignore this email.
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`
)

type SendEmailRequest struct {
	RecipientEmail   string `json:"recipient_email"`
	RecipientName    string `json:"recipient_name"`
	VerificationCode string `json:"verification_code"`
	Subject          string `json:"subject"`
}

func SendEmailVerificationCode(recipientEmail string, recipientName string, verificationCode string, companyName string) error {
	htmlBody := strings.Replace(EmailVerificationTemplate, "{{name}}", recipientName, 1)
	htmlBody = strings.Replace(htmlBody, "{{verification_code}}", verificationCode, 1)
	htmlBody = strings.Replace(htmlBody, "{{year}}", strconv.Itoa(time.Now().Year()), 1)
	htmlBody = strings.Replace(htmlBody, "{{company_name}}", companyName, 1)

	err := SendRawEmail(recipientEmail, "Email Verification", htmlBody)
	if err != nil {
		return err
	}

	/* err := SendEmail(recipientEmail, "Email Verification", htmlBody)
	if err != nil {
		return err
	} */

	return nil
}

func SendPasswordResetEmail(recipientEmail string, recipientName string, verificationCode string, companyName string) error {
	htmlBody := strings.Replace(PasswordResetTemplate, "{{name}}", recipientName, 1)
	htmlBody = strings.Replace(htmlBody, "{{verification_code}}", verificationCode, 1)
	htmlBody = strings.Replace(htmlBody, "{{year}}", strconv.Itoa(time.Now().Year()), 1)
	htmlBody = strings.Replace(htmlBody, "{{company_name}}", companyName, 1)

	err := SendRawEmail(recipientEmail, "Password Reset", htmlBody)
	if err != nil {
		return err
	}

	return nil
}

func SendEmail(recipientEmail string, subject string, htmlBody string) error {
	mail := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{aws.String(recipientEmail)},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Data: aws.String(subject),
			},
			Body: &ses.Body{
				Html: &ses.Content{
					Data: aws.String(htmlBody),
				},
			},
		},
		Source: aws.String(awsConfig.Sender),
	}

	_, err := SESClient.SendEmail(mail)
	if err != nil {
		return err
	}

	return nil
}

func SendRawEmail(recipientEmail string, subject string, htmlBody string) error {
	// Create the raw email message
	rawMessage, err := createRawEmailMessage(recipientEmail, subject, htmlBody)
	if err != nil {
		return err
	}

	// Create SendRawEmailInput
	mail := &ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{
			Data: rawMessage,
		},
		//Source: aws.String(Sender),
		Source: aws.String(awsConfig.Sender),
		Destinations: []*string{
			aws.String(recipientEmail),
		},
	}

	_, err = SESClient.SendRawEmail(mail)
	if err != nil {
		return err
	}

	return nil
}

func createRawEmailMessage(recipientEmail string, subject string, htmlBody string) ([]byte, error) {
	// Create a buffer to hold the email
	var buffer bytes.Buffer

	// Create the email headers with proper formatting

	headers := make(map[string]string)
	headers["From"] = awsConfig.Sender
	headers["To"] = recipientEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Date"] = time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700")

	// Write headers
	for key, value := range headers {
		buffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	buffer.WriteString("\r\n")

	// Write the HTML body
	buffer.WriteString(htmlBody)
	fmt.Println(htmlBody)
	return buffer.Bytes(), nil
}

// test api send email
//
//encore:api public method=POST path=/api/aws-ses/send-email-test
func (s *Service) SendEmaiTestAPI(ctx context.Context, req SendEmailRequest) error {
	return SendEmailVerificationCode(
		req.RecipientEmail,
		req.RecipientName,
		req.VerificationCode,
		"MRCHURROS",
	)
}
