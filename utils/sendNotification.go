package utils

import (
	"context"
	"fmt"
	"laughifi/database"

	"firebase.google.com/go/messaging"
)

var ctx = context.Background()

func SendNotificationToUser(
	deviceToken string,
	deviceType string,
	title string,
	body string,
	data map[string]string,
) error {

	var message *messaging.Message = &messaging.Message{
		Data: data,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: deviceToken,
	}

	_, err := database.GetFirebaseMessagingClient().Send(ctx, message)
	if err != nil {
		return fmt.Errorf("Error sending notification: %v\n", err)
	}

	return nil
}
