package websocket

import (
	"encoding/json"
	"laughifi/database"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type DisConnData struct {
	ConnectionID string `json:"connectionId"`
}

func WebsocketDisconnect(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var disConnData DisConnData
		if err := json.NewDecoder(r.Body).Decode(&disConnData); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		filter := bson.M{
			"connectionId": disConnData.ConnectionID,
		}

		_, err := db.GetCollection("websocketConnection").DeleteOne(ctx, filter)
		if err != nil {
			http.Error(w, "Failed to delete connection", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}
}
