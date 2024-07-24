package utils

import (
	"context"
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
) {

	var message *messaging.Message = &messaging.Message{
		Data: data,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: deviceToken,
	}

	database.GetFirebaseMessagingClient().Send(ctx, message)
}
