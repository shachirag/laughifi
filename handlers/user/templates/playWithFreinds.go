package templates

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
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
			ShareIds:   []primitive.ObjectID{user.Id, friendObjID},
			TemplateId: templateObjId,
			Answers:    answers,
			IsFriend:   true,
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

	friendDetails, err := fetchFriendDetails(ctx, db, friendObjIDs)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch friend details: %v", err)
	}

	title := "Template Shared"
	body := fmt.Sprintf("%s has shared a template with you", user.Name)
	notifData := map[string]string{
		"templateId": templateObjId.Hex(),
		"type":       "templateShared",
	}

	errCh := make(chan error, len(friendObjIDs))

	for _, friendObjID := range friendObjIDs {
		if friendObjID != user.Id {
			go func(friendID primitive.ObjectID) {
				friend := friendDetails[friendID.Hex()]

				if friend.DeviceInfo.DeviceToken != "" && friend.DeviceInfo.DeviceType != "" {
					utils.SendNotificationToUser(friend.DeviceInfo.DeviceToken, friend.DeviceInfo.DeviceType, title, body, notifData)
				}
			}(friendObjID)
		}
	}

	for i := 0; i < len(friendObjIDs)-1; i++ {
		select {
		case err := <-errCh:
			if err != nil {
				return nil, gqlerror.Errorf(err.Error())
			}
		}
	}

	return &model.Response{
		Message: "Successfully Shared",
	}, nil
}

func fetchFriendDetails(ctx context.Context, db *database.DB, friendObjIDs []primitive.ObjectID) (map[string]entity.CustomerEntity, error) {
	filter := bson.M{"_id": bson.M{"$in": friendObjIDs}}
	cursor, err := db.GetCollection("customer").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	friendDetails := make(map[string]entity.CustomerEntity)
	for cursor.Next(ctx) {
		var friend entity.CustomerEntity
		if err := cursor.Decode(&friend); err != nil {
			return nil, err
		}
		friendDetails[friend.Id.Hex()] = friend
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return friendDetails, nil
}
