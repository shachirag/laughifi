package utils

import (
	"fmt"
	"os"
	"strings"
)

func SendForgotPasswordEmail(userEmail string, userName string, otp string) error {
	emailBody := getEmailData()
	emailBody = strings.ReplaceAll(emailBody, "[APP_URL]", os.Getenv("AWS_S3_BUCKET_URL"))
	emailBody = strings.ReplaceAll(emailBody, "[User Name]", userName)
	emailBody = strings.ReplaceAll(emailBody, "[Generated OTP]", otp)
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
		Subject:  "OTP for reset password.",
		HTMLBody: emailBody,
	}

	err := SendTransactionalEmail(emailData)
	if err != nil {
		return fmt.Errorf("Failed to send email: %s", err)
	}

	return nil
}

func getEmailData() string {
	return `<!DOCTYPE html>
	<html lang="en">
	
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
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
				background-color: #4CAF50;
				color: white;
			}
		</style>
	</head>
	
	<body>
		<table>
			<tr>
				<th colspan="2" style="text-align: left; padding: 15px; background-color: #fff; color: white;">
					<a href="#"><img src="[APP_URL]/abbsi/512_512+-round.png" alt=""></a>
				</th>
			</tr>
			<tr>
				<td style="border: none;">
					<h2>ABBSI Admin Panel - Password Recovery OTP</h2>
				</td>
			</tr>
			<tr>
				<td colspan="2" style="padding: 15px;">
					<p>Dear [User Name],</p>
					<p>It appears that a password reset request has been initiated for your ABBSI Admin Panel account. To
						continue with the password recovery process, a One-Time Password (OTP) has been generated for you.
					</p>
					<p><b>OTP:</b>
						[Generated OTP]</p>
					<p>This OTP is valid for a limited time to ensure the security of your account. If you did not request a
						password recovery or if you have any concerns, please contact our support team immediately at
						[Support Email or Phone Number].</p>
	
					<p>Thank you for your prompt attention to this matter. We appreciate your dedication to maintaining a
						secure and efficient administration on the ABBSI platform.</p>
					<p>Best regards,<br>
						ABBSI<br>
				</td>
			</tr>
			<tr>
				<td colspan="2" style="text-align: center; padding: 15px; background-color: #f4f4f4;">
					<p>This is a notification email from <b>ABBSI</b>. Please do not reply to this email.</p>
				</td>
			</tr>
		</table>
	</body>
	
	</html>`
}
