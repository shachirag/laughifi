package utils

import (
	"fmt"
	"os"
	"strings"
)

func SendSignupEmail(userEmail string, userName string, otp string) error {
	emailBody := getSignupEmailData()
	emailBody = strings.ReplaceAll(emailBody, "[Recipient's Name]", userName)
	emailBody = strings.ReplaceAll(emailBody, "[OTP Code]", otp)
	senderEmail := os.Getenv("SENDER_EMAIL")
	emailData := Email{
		From: EmailFrom{
			Address: senderEmail,
			Name:    "Laughifi",
		},
		To: []EmailTo{{
			EmailAddress: EmailAddress{
				Address: userEmail,
			},
		}},
		Subject:  "OTP for Signup.",
		HTMLBody: emailBody,
	}

	err := SendTransactionalEmail(emailData)
	if err != nil {
		return fmt.Errorf("Failed to send email: %s", err)
	}

	return nil
}

func getSignupEmailData() string {
	return `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title></title>
    <style>
      body {
        font-family: Arial, sans-serif;
        line-height: 1.6;
        margin: 0;
        padding: 0;
      }

      table {
        width: 100%;
        max-width: 600px;
        margin: 0 auto;
      }

      th,
      td {
        padding: 10px;
        text-align: left;
        border-bottom: 1px solid #ddd;
      }

      th {
        background-color: #4caf50;
        color: white;
      }
    </style>
  </head>

  <body>
    <table>
      <tr>
        <td style="border: none">
          <h2 style="margin-bottom: 0">Reset Your Laughify Password</h2>
        </td>
      </tr>
      <tr>
        <td colspan="2" style="padding: 10px">
          <p>Dear [Recipient's Name],</p>
          <p>We received a request to reset your Laughify password.</p>
          <p>
            Please use the One-Time Password (OTP) provided below to reset your password:
          </p>
          <p>Your OTP Code: [OTP Code]</p>
          <p>
            This code is valid for the next 10 minutes.
          </p>
          <p>
            If you did not request a password reset, please ignore this email or contact our support team immediately.
          </p>
          <p>
            Best Regards,<br />
            <b>The Laughify Team</b>
          </p>
        </td>
      </tr>
      <tr>
        <td
          colspan="2"
          style="text-align: center; padding: 15px; background-color: #f4f4f4"
        >
          <p>
            This is a notification email from <b>The Laughify Team</b>. Please do not
            reply to this email.
          </p>
        </td>
      </tr>
    </table>
  </body>
</html>
`
}
