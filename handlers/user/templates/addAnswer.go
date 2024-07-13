package templates

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils/websocket"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddAnswer(ctx context.Context, db *database.DB, data model.AddAnswerRequestInput, id string) (*model.AddAnswerResponse, error) {
	var (
		templatePlayWithFriendColl = db.GetCollection("templatePlayWithFriend")
		playWithFriendTemplate     entity.TemplatePlayWithFriendEntity
	)

	playWithFriendTemplateObjId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid template Id")
	}

	sendedUserObjId, err := primitive.ObjectIDFromHex(data.SendAnswerUserID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid friend Id")
	}

	filter := bson.M{
		"_id": playWithFriendTemplateObjId,
	}

	err = templatePlayWithFriendColl.FindOne(ctx, filter).Decode(&playWithFriendTemplate)
	if err != nil {
		return nil, gqlerror.Errorf("Internal server error while fetching the template.")
	}

	templateFilter := bson.M{
		"_id": playWithFriendTemplate.TemplateId,
	}
	var template entity.TemplatesEntity
	err = db.GetCollection("template").FindOne(ctx, templateFilter).Decode(&template)
	if err != nil {
		return nil, gqlerror.Errorf("Internal server error while fetching the template.")
	}

	if len(playWithFriendTemplate.Answers) > 0 {
		lastAnswer := playWithFriendTemplate.Answers[len(playWithFriendTemplate.Answers)-1]
		if lastAnswer.Id == sendedUserObjId {
			return &model.AddAnswerResponse{
				Message: "It,s not your turn to answer",
				Status:  false,
			}, nil
		}
	}

	newAnswer := entity.AnswersTemplatePlayWithFriendEntity{
		Id:    sendedUserObjId,
		Key:   data.Key,
		Value: data.Value,
	}

	playWithFriendTemplate.Answers = append(playWithFriendTemplate.Answers, newAnswer)

	update := bson.M{
		"$push": bson.M{"answers": newAnswer},
		"$set":  bson.M{"updatedAt": time.Now().UTC()},
	}

	_, err = templatePlayWithFriendColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to add answer")
	}

	requiredPlaceholders := extractPlaceholders(template.Template)
	filledPlaceholders := make(map[string]bool)
	for _, answer := range playWithFriendTemplate.Answers {
		filledPlaceholders[answer.Key] = true
	}

	allFilled := true
	for _, placeholder := range requiredPlaceholders {
		if !filledPlaceholders[placeholder] {
			allFilled = false
			break
		}
	}

	if allFilled {
		update := bson.M{
			"$set": bson.M{"status": "completed"},
		}
		_, err = templatePlayWithFriendColl.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to update status to completed")
		}
	}

	var secondLastUserObjId primitive.ObjectID
	if len(playWithFriendTemplate.Answers) > 0 {
		secondLastUser := playWithFriendTemplate.Answers[len(playWithFriendTemplate.Answers)-2]
		secondLastUserObjId = secondLastUser.Id
	}

	fmt.Println("secondLastUser", secondLastUserObjId)
	websocket.SendTemplateData(db, playWithFriendTemplateObjId, secondLastUserObjId)

	return &model.AddAnswerResponse{
		Message: "Success",
		Status:  true,
	}, nil
}

func extractPlaceholders(template string) []string {
	var placeholders []string
	parts := strings.Split(template, "{{")
	for _, part := range parts[1:] {
		placeholder := strings.Split(part, "}}")[0]
		placeholders = append(placeholders, strings.TrimSpace(placeholder))
	}
	return placeholders
}
