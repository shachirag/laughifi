package templates

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TemplatePlayWithFriends(ctx context.Context, db *database.DB, data model.TemplatePlayWithFriendsRequestInput) (*model.Response, error) {
	var (
		templatePlayWithFriendColl = db.GetCollection("templatePlayWithFriend")
	)

	templateObjId, err := primitive.ObjectIDFromHex(data.TemplateID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid template Id")
	}

	friendIdStrs := strings.Split(data.FriendIds, ",")
	var friendObjIDs []primitive.ObjectID
	for _, friendIdStr := range friendIdStrs {
		friendObjID, err := primitive.ObjectIDFromHex(friendIdStr)
		if err != nil {
			return nil, gqlerror.Errorf("invalid friend Id")
		}
		friendObjIDs = append(friendObjIDs, friendObjID)
	}

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	if data.Answer == nil || len(data.Answer) == 0 {
		return nil, gqlerror.Errorf("At least one answer is required to share")
	}

	var answers []entity.AnswersTemplatePlayWithFriendEntity
	if data.Answer != nil {
		for _, answerInput := range data.Answer {
			answer := entity.AnswersTemplatePlayWithFriendEntity{
				Id:    user.Id,
				Key:   answerInput.Key,
				Value: answerInput.Value,
			}
			answers = append(answers, answer)
		}
	}

	var insertDocuments []interface{}
	for _, friendObjID := range friendObjIDs {
		id := primitive.NewObjectID()

		templatePlayWithFriend := entity.TemplatePlayWithFriendEntity{
			Id:         id,
			UserId:     user.Id,
			TemplateId: templateObjId,
			FriendId:   friendObjID,
			Answers:    answers,
			Status:     "pending",
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}

		insertDocuments = append(insertDocuments, templatePlayWithFriend)

	}

	_, err = templatePlayWithFriendColl.InsertMany(ctx, insertDocuments)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to share templates with friends: %v", err)
	}

	return &model.Response{
		Message: "Successfully Shared",
	}, nil
}
