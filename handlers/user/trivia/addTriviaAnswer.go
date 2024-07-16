package trivia

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddTriviaAnswer(ctx context.Context, db *database.DB, ownGameId string, answer string) (*model.Response, error) {
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
				"playedStatus": "approved",
			},
		},
	}

	var ownGame entity.OwnGameEntity
	err = db.GetCollection("ownGame").FindOne(ctx, filter).Decode(&ownGame)
	if err != nil {
		return nil, gqlerror.Errorf("Error occurred while fetching own game: " + err.Error())
	}

	var triviaId primitive.ObjectID
	if ownGame.TriviaID != nil {
		triviaId = *ownGame.TriviaID
	}

	var trivia entity.TriviaEntity
	triviaFilter := bson.M{
		"_id": triviaId,
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

	return &model.Response{
		Message: "Answer Added",
	}, nil
}
