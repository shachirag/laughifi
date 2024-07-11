package templates

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils/subscription"
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

	fmt.Printf("Notifying subscribers with template ID: %s\n", sharedTemplate.ID)
	subscription.NewManager().NotifySubscribers(sharedTemplate)

	return &model.Response{
		Message: "Success",
	}, nil
}
