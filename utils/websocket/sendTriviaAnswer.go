package websocket

import (
	"encoding/json"
	"fmt"
	"laughifi/database"
	"laughifi/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SendTriviaData(db *database.DB, ownGameId primitive.ObjectID, customerId primitive.ObjectID) {

	if ownGameId.IsZero() {
		return
	}

	newTriviaDto := GetTrivaiAnswerDto(db, ownGameId)
	if newTriviaDto == nil {
		return
	}

	dataMap := map[string]interface{}{
		"action":  "trivia",
		"payload": newTriviaDto,
	}

	dataBytes, err := json.Marshal(dataMap)
	if err != nil {
		return
	}

	if dataBytes == nil {
		return
	}

	filter := bson.M{
		"roleId": customerId,
	}

	abc, _ := json.Marshal(filter)
	fmt.Println(string(abc))

	cur, err := db.GetCollection("websocketConnection").Find(ctx, filter)
	if err == nil {
		defer cur.Close(ctx)
		websocketConnectionEntities := []entity.WebsocketConnectionEntity{}
		err = cur.All(ctx, &websocketConnectionEntities)
		if err == nil && len(websocketConnectionEntities) > 0 {
			for _, ws := range websocketConnectionEntities {
				err := sendConnectionIDToAPIGateway(db, ws.ConnectionId, &dataBytes)
				if err != nil {
					fmt.Println("63", err)
				}
			}
		}
	}
}

