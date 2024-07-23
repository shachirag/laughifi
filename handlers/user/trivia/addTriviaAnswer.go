package trivia

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"laughifi/utils/websocket"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddTriviaAnswer(ctx context.Context, db *database.DB, ownGameId string, answer string, role string) (*model.Response, error) {
	var (
		ownGameColl = db.GetCollection("ownGame")
	)

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	ownGameObjID, err := primitive.ObjectIDFromHex(ownGameId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid own game ID")
	}

	filter := bson.M{
		"_id": ownGameObjID,
		"friends": bson.M{
			"$elemMatch": bson.M{
				"id":           user.Id,
				"playedStatus": "accepted",
				"role":         role,
			},
		},
	}

	var ownGame entity.OwnGameEntity
	err = db.GetCollection("ownGame").FindOne(ctx, filter).Decode(&ownGame)
	if err != nil {
		return nil, gqlerror.Errorf("Error occurred while fetching own game: " + err.Error())
	}

	for _, friend := range ownGame.Friends {
		if friend.Id == user.Id && friend.Answer != nil {
			return nil, gqlerror.Errorf("You have already submitted an answer")
		}
	}

	var trivia entity.TriviaEntity
	triviaFilter := bson.M{
		"_id": ownGame.TriviaID,
	}

	err = db.GetCollection("trivia").FindOne(ctx, triviaFilter).Decode(&trivia)
	if err != nil {
		return nil, gqlerror.Errorf("Error occurred while fetching trivia: " + err.Error())
	}

	isCorrectAnswer := false
	if answer == trivia.CorrectAnswer {
		isCorrectAnswer = true
	}

	update := bson.M{
		"$set": bson.M{
			"friends.$.answer":          answer,
			"friends.$.isCorrectAnswer": isCorrectAnswer,
			"updatedAt":                 time.Now().UTC(),
		},
	}

	_, err = ownGameColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to upadte answer")
	}

	err = db.GetCollection("ownGame").FindOne(ctx, bson.M{"_id": ownGameObjID}).Decode(&ownGame)
	if err != nil {
		return nil, gqlerror.Errorf("Error occurred while fetching updated own game: " + err.Error())
	}

	allAnswered := true
	for _, friend := range ownGame.Friends {
		if friend.Answer == nil {
			allAnswered = false
			break
		}
	}

	status := "pending"
	if allAnswered {
		status = "answered"
	}

	statusUpdate := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now().UTC(),
		},
	}

	_, err = ownGameColl.UpdateOne(ctx, bson.M{"_id": ownGameObjID}, statusUpdate)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update own game status")
	}

	friendObjIDs := make([]primitive.ObjectID, len(ownGame.Friends))
	for i, friend := range ownGame.Friends {
		friendObjIDs[i] = friend.Id
	}

	friendDetails, err := fetchFriendDetails(ctx, db, friendObjIDs)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch friend details: %v", err)
	}

	title := "Trivia Game Update"
	body := fmt.Sprintf("%s has answered the trivia game %s.", user.Name, ownGame.GameName)
	notifData := map[string]string{
		"type":   "triviaAnswer",
		"gameId": ownGameObjID.Hex(),
	}

	// for _, friend := range ownGame.Friends {
	// 	// Skip sending notification to the user who answered
	// 	if friend.Id == user.Id {
	// 		continue
	// 	}

	// 	if friendDetails[friend.Id.Hex()].DeviceInfo.DeviceToken != "" && friendDetails[friend.Id.Hex()].DeviceInfo.DeviceType != "" {
	// 		friendDetails := friendDetails[friend.Id.Hex()]
	// 		err := utils.SendNotificationToUser(friendDetails.DeviceInfo.DeviceToken, friendDetails.DeviceInfo.DeviceType, title, body, notifData)
	// 		if err != nil {
	// 			return nil, gqlerror.Errorf("Failed to send notification to friend: %v", err)
	// 		}
	// 	}
	// }

	errCh := make(chan error, len(ownGame.Friends))

	for _, friend := range ownGame.Friends {
		if friend.Id == user.Id {
			continue
		}

		go func(friendID primitive.ObjectID) {
			if friendDetails[friendID.Hex()].DeviceInfo.DeviceToken != "" && friendDetails[friendID.Hex()].DeviceInfo.DeviceType != "" {
				friendDetails := friendDetails[friendID.Hex()]
				err := utils.SendNotificationToUser(friendDetails.DeviceInfo.DeviceToken, friendDetails.DeviceInfo.DeviceType, title, body, notifData)
				if err != nil {
					errCh <- err
				}
			}
		}(friend.Id)
	}

	for i := 0; i < len(ownGame.Friends)-1; i++ {
		if err := <-errCh; err != nil {
			return nil, gqlerror.Errorf("Failed to send notification to friend: %v", err)
		}
	}

	// if status == "answered" {
	// 	title := "Trivia Game Update"
	// 	body := fmt.Sprintf("All answers have been submitted for %s", ownGame.GameName)
	// 	data := map[string]string{
	// 		"gameId": ownGameObjID.Hex(),
	// 		"type":   "allTriviaAnswerSubmitted",
	// 	}

	// 	for _, friend := range ownGame.Friends {
	// 		err := utils.SendNotificationToUser(friend.DeviceInfo.DeviceToken, friend.DeviceInfo.DeviceType, title, body, data)
	// 		if err != nil {
	// 			return nil, gqlerror.Errorf("Failed to send notification: %s", err.Error())
	// 		}
	// 	}
	// }

	if status == "answered" {
		title := "Trivia Game Update"
		body := fmt.Sprintf("All answers have been submitted for %s", ownGame.GameName)
		notifData := map[string]string{
			"gameId": ownGameObjID.Hex(),
			"type":   "allTriviaAnswerSubmitted",
		}
		for _, friend := range ownGame.Friends {
			if friendDetails[friend.Id.Hex()].DeviceInfo.DeviceToken != "" && friendDetails[friend.Id.Hex()].DeviceInfo.DeviceType != "" {
				friendDetails := friendDetails[friend.Id.Hex()]
				err := utils.SendNotificationToUser(friendDetails.DeviceInfo.DeviceToken, friendDetails.DeviceInfo.DeviceType, title, body, notifData)
				if err != nil {
					return nil, gqlerror.Errorf("Failed to send notification to friend: %v", err)
				}
			}
		}

		errCh := make(chan error, len(ownGame.Friends))

		for _, friend := range ownGame.Friends {
			if friend.Id == user.Id {
				continue
			}

			go func(friendID primitive.ObjectID) {
				if friendDetails[friendID.Hex()].DeviceInfo.DeviceToken != "" && friendDetails[friendID.Hex()].DeviceInfo.DeviceType != "" {
					friendDetails := friendDetails[friendID.Hex()]
					err := utils.SendNotificationToUser(friendDetails.DeviceInfo.DeviceToken, friendDetails.DeviceInfo.DeviceType, title, body, notifData)
					if err != nil {
						errCh <- err
					}
				}
			}(friend.Id)
		}

		for i := 0; i < len(ownGame.Friends)-1; i++ {
			if err := <-errCh; err != nil {
				return nil, gqlerror.Errorf("Failed to send notification to friend: %v", err)
			}
		}
	}

	for _, friend := range ownGame.Friends {
		websocket.SendTriviaData(db, ownGameObjID, friend.Id)
	}

	return &model.Response{
		Message: "Answer Added",
	}, nil
}
