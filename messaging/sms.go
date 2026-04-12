package messaging

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func SendSMS(phoneNumber string, message string, user string, pass string) error {
	function := "[messaging.SendSMS]"

	params := url.Values{}
	params.Set("to", phoneNumber)
	params.Set("text", message)
	params.Set("user", user)
	params.Set("pass", pass)

	baseURL := "https://api.anchor-sms.com:8443/mt/sendMtSmsNoToken"

	fullURL := baseURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("%s error creating request: %v", function, err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s error sending request: %v", function, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%s failed to read response body: %v", function, err)
		}

		return fmt.Errorf("%s failed to send SMS: %s, %s", function, resp.Status, string(body))
	}

	return nil
}
