package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ctx = context.Background()

func SendTemplateData(db *database.DB, templateId primitive.ObjectID, customerId primitive.ObjectID) {

	if templateId.IsZero() {
		return
	}

	newTemplateDto := GetTemplateDto(db, templateId)
	if newTemplateDto == nil {
		return
	}

	dataMap := map[string]interface{}{
		"action":  "template",
		"payload": newTemplateDto,
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

	// abc, _ := json.Marshal(filter)
	// fmt.Println(string(abc))

	cur, err := db.GetCollection("websocketConnection").Find(ctx, filter)
	if err == nil {
		defer cur.Close(ctx)
		websocketConnectionEntities := []entity.WebsocketConnectionEntity{}
		err = cur.All(ctx, &websocketConnectionEntities)
		if err == nil && len(websocketConnectionEntities) > 0 {
			for _, ws := range websocketConnectionEntities {
				err := sendConnectionIDToAPIGateway(db, ws.ConnectionId, &dataBytes)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func sendConnectionIDToAPIGateway(db *database.DB, connectionID string, data *[]byte) error {
	input := &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         *data,
	}

	_, err := database.GetApiClient().PostToConnection(ctx, input)
	if err != nil {
		if strings.Contains(err.Error(), "StatusCode: 410") {
			filter := bson.M{"connectionId": connectionID}
			db.GetCollection("websocketConnection").DeleteOne(ctx, filter)
		}
		return err
	}
	return nil
}
