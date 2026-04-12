package messaging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const BaseURL = "https://graph.facebook.com"
const APIVersion = "v22.0"
const BusinessPhoneNumberID = "916783011526590"
const tempAccessToken = "EAATVCqee5NsBQTYr07oi87jZAcGw5fvLGXWQGSZBWx9OI9fE6mX1chg19jwZBC3pZASZAOreKlv1l7qZAwaD9AdoDUpCgbZBr1SXtZAi71qpNRKWybvgHWbnhzmAKvBLr3ZBecb7cZBXPM3d8T3DjtCUzqMiS5eFkB5gesiMT1LP6FDxecq7UBAcNuLeOl2UYCGGNjE11AMSSvDEbDe0iqGMKeZAE8BjDiTsj1UZCLtFcDv20UebZAdZANedy4oG6PZCZCAZBq4pITtDZAxQlPLekbvTY8bhX3AtSebgZDZD"

func SendZeroTapAuthCode() error {
	requestBody := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                "60189500482",
		"type":              "template",
		"template": map[string]interface{}{
			"name": "verification_code",
			"language": map[string]interface{}{
				"code": "en_US",
			},
			"components": []map[string]interface{}{
				{
					"type": "body",
					"parameters": []map[string]interface{}{
						{
							"type": "text",
							"text": "123456",
						},
					},
				},
				{
					"type":     "button",
					"sub_type": "url",
					"index":    "0",
					"parameters": []map[string]interface{}{
						{
							"type": "text",
							"text": "123456",
						},
					},
				},
			},
		},
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf("%s/%s/%s/messages", BaseURL, APIVersion, BusinessPhoneNumberID)

	httpReq, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+tempAccessToken)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %s", err)
		}
		return fmt.Errorf("failed to send message: %s, response body: %s", resp.Status, string(body))
	}

	return nil
}
