package templates

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils/subscription"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddAnswer(ctx context.Context, db *database.DB, data model.AddAnswerRequestInput, id string) (*model.Response, error) {
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
			return nil, gqlerror.Errorf("It's not your turn to answer.")
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

	sharedTemplate := &model.TemplatePlayWithFriend{
		ID:       playWithFriendTemplate.Id.Hex(),
		FriendID: playWithFriendTemplate.FriendId.Hex(),
		Template: template.Template,
		Topic:    template.Category.Name,
		Title:    template.Title,
		Status:   playWithFriendTemplate.Status,
		Answers:  make([]*model.TemplateAnswers, len(playWithFriendTemplate.Answers)),
	}

	for i, answer := range playWithFriendTemplate.Answers {
		sharedTemplate.Answers[i] = &model.TemplateAnswers{
			ID:    answer.Id.Hex(),
			Key:   answer.Key,
			Value: answer.Value,
		}
	}

	subManager := subscription.NewManager()

	secondLastIndex := len(playWithFriendTemplate.Answers) - 2
	if secondLastIndex >= 0 {
		secondLastUserID := playWithFriendTemplate.Answers[secondLastIndex].Id.Hex()
		if subManager.SubscriberExists(secondLastUserID) {
			if err := subManager.NotifySubscriberByID(sharedTemplate, secondLastUserID); err != nil {
				fmt.Printf("Error notifying subscriber: %s\n", err)
			}
		} else {
			fmt.Printf("129 Subscriber with ID %s does not exist\n", secondLastUserID)
		}
	}
	// fmt.Printf("Notifying subscribers with template ID: %s\n", sharedTemplate.ID)
	// subscription.NewManager().NotifySubscribersByID(sendedUserObjId)

	return &model.Response{
		Message: "Success",
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
