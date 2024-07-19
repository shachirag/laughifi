package websocket

import (
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetTrivaiAnswerDto(db *database.DB, ownGameId primitive.ObjectID) *model.TriviaDetail {

	var ownGame entity.OwnGameEntity

	filter := bson.M{
		"_id": ownGameId,
	}

	err := db.GetCollection("ownGame").FindOne(ctx, filter).Decode(&ownGame)
	if err != nil {
		return nil
	}

	var trivia entity.TriviaEntity

	triviaFilter := bson.M{
		"_id": ownGame.TriviaID,
	}

	err = db.GetCollection("trivia").FindOne(ctx, triviaFilter).Decode(&trivia)
	if err != nil {
		return nil
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
		} else if friend.PlayedStatus == "pending" {
			notAnsweredUsers = append(notAnsweredUsers, user)
		}
	}

	newTemplateObj := model.TriviaDetail{
		ID:                 ownGame.Id.Hex(),
		Question:           trivia.Question,
		Answers:            trivia.Answers,
		AnsweredUsers:      answeredUsers,
		NotAnsweredUsers:   notAnsweredUsers,
		DareForWrongAnswer: dareForWrongAnswer,
		Status:             ownGame.Status,
		CorrectAnswer:      correctAnswer,
	}

	return &newTemplateObj
}
