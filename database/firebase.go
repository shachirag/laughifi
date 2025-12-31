package database

import (
	"log"
	"os"
	"path"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

var (
	firebaseApp     *firebase.App
	messagingClient *messaging.Client
)

func StartFirebase() error {
	wd, _ := os.Getwd()
	saPath := ""
	// saPath := path.Join(wd, "firebase-sa-creds.json")
	sa := option.WithCredentialsFile(saPath)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatal("Error creating a firebase app instance ", err)
	}

	messagingClient, err = app.Messaging(ctx)
	if err != nil {
		log.Fatal("Error creating messaging client", err)
	}

	return nil
}

func GetFirebaseMessagingClient() *messaging.Client {
	return messagingClient
}
