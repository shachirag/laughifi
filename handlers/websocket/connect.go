package websocket

import (
	"context"
	"encoding/json"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/middleware"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ConnectionData struct {
	ConnectionID string             `json:"connectionId"`
	Token        string             `json:"token"`
	Id           primitive.ObjectID `json:"id" bson:"id"`
	Type         string             `json:"type" bson:"type"`
}

var ctx = context.Background()

func WebsocketConnect(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var connData ConnectionData
		if err := json.NewDecoder(r.Body).Decode(&connData); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		token := connData.Token
		data, err := middleware.SetWebSocketMiddlewareData(token)
		if err != nil {
			http.Error(w, "Failed to authenticate token", http.StatusInternalServerError)
			return
		}

		webconnection := entity.WebsocketConnectionEntity{
			Id:           primitive.NewObjectID(),
			ConnectionId: connData.ConnectionID,
			RoleId:       data.ID,
			Role:         data.Role,
		}

		switch connData.Type {
		case "trivia":
			webconnection.GameType = &connData.Type
			webconnection.TriviaId = &connData.Id
		case "template":
			webconnection.GameType = &connData.Type
			webconnection.TemplateId = &connData.Id
		}

		collection := db.GetCollection("websocketConnection")

		if data.Role == "customer" {
			filter := bson.M{"role": data.Role, "roleId": data.ID}
			update := bson.M{
				"$set": bson.M{
					"connectionId": connData.ConnectionID,
				},
			}

			result := collection.FindOne(ctx, filter)
			if result.Err() != nil {
				if result.Err() == mongo.ErrNoDocuments {
					_, err = collection.InsertOne(ctx, webconnection)
					if err != nil {
						http.Error(w, "Failed to insert new connection", http.StatusInternalServerError)
						return
					}
				} else {
					http.Error(w, "Failed to find connection", http.StatusInternalServerError)
					return
				}
			} else {
				_, err = collection.UpdateOne(ctx, filter, update)
				if err != nil {
					http.Error(w, "Failed to update connection", http.StatusInternalServerError)
					return
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}
}
