package trivia

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTriviaDetail(ctx context.Context, db *database.DB, id string) (*model.TriviaDetail, error) {
	var ownGame entity.OwnGameEntity

	ownGameObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid trivia Id")
	}

	ownGameColl := db.GetCollection("ownGame")

	err = ownGameColl.FindOne(ctx, bson.M{"_id": ownGameObjID}).Decode(&ownGame)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Own Game not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch Own Game")
	}

	var trivia entity.TriviaEntity
	err = db.GetCollection("trivia").FindOne(ctx, bson.M{"_id": ownGame.TriviaID}).Decode(&trivia)
	if err != nil {
		return nil, gqlerror.Errorf("Internal server error while fetching the trivia: " + err.Error())
	}

	var answeredUsers []*model.Users
	var notAnsweredUsers []*model.Users
	var dareForWrongAnswer *string
	var correctAnswer *string

	if ownGame.Status == "answered" {
		dareForWrongAnswer = &trivia.DareForWrongAnswer
		correctAnswer = &trivia.CorrectAnswer
	}

	for _, friend := range ownGame.Friends {
		user := &model.Users{
			ID:    friend.Id.Hex(),
			Name:  friend.Name,
			Image: friend.Image,
		}

		if friend.Answer != nil {
			user.Answer = friend.Answer
		}

		if ownGame.Status == "answered" && friend.IsCorrectAnswer != nil {
			user.IsCorrectAnswer = friend.IsCorrectAnswer
		}

		if friend.Answer != nil {
			answeredUsers = append(answeredUsers, user)
		} else if friend.PlayedStatus == "accepted" {
			notAnsweredUsers = append(notAnsweredUsers, user)
		}
	}

	return &model.TriviaDetail{
		ID:                 ownGame.Id.Hex(),
		Question:           trivia.Question,
		Answers:            trivia.Answers,
		AnsweredUsers:      answeredUsers,
		NotAnsweredUsers:   notAnsweredUsers,
		DareForWrongAnswer: dareForWrongAnswer,
		Status:             ownGame.Status,
		CorrectAnswer:      correctAnswer,
	}, nil
}
