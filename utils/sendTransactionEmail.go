package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func SendTransactionalEmail(emailData Email) error {
	jsonData, err := json.Marshal(emailData)
	if err != nil {
		return fmt.Errorf("Failed to marshal email data: %s", err)
	}

	req, err := http.NewRequest("POST", "https://api.zeptomail.com/v1.1/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Failed to create HTTP request: %s", err)
	}

	zeptomailToken := os.Getenv("ZEPTOEMAIL_TOKEN")

	req.Header.Set("Authorization", "Zoho-enczapikey "+zeptomailToken)

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send HTTP request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Received non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

type Email struct {
	From EmailFrom `json:"from"`
	To []EmailTo`json:"to"`
	Subject  string `json:"subject"`
	HTMLBody string `json:"htmlbody"`
}

type EmailFrom struct {
	Address string `json:"address"`
	Name string `json:"name"`
}

type EmailTo struct {
	EmailAddress EmailAddress `json:"email_address"`
}

type EmailAddress struct {
	Address string `json:"address"`
}
