package trivia

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"math/rand"
	"time"

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

	var trivias []entity.TriviaEntity
	triviaColl := db.GetCollection("trivia")
	cursor, err := triviaColl.Find(ctx, bson.M{"category.id": ownGame.Category.Id})
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch trivia")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var trivia entity.TriviaEntity
		if err := cursor.Decode(&trivia); err != nil {
			return nil, gqlerror.Errorf("Failed to decode trivia")
		}
		trivias = append(trivias, trivia)
	}

	if len(trivias) == 0 {
		return nil, gqlerror.Errorf("Trivia not Found")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomTrivia := trivias[r.Intn(len(trivias))]

	update := bson.M{
		"$set": bson.M{
			"triviaId":  randomTrivia.Id,
			"updatedAt": time.Now().UTC(),
		},
	}

	_, err = ownGameColl.UpdateOne(ctx, bson.M{"_id": ownGameObjID}, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update Own Game with trivia ID")
	}

	var answeredUsers []*model.Users
	var notAnsweredUsers []*model.Users
	var dareForWrongAnswer *string

	if ownGame.Status == "answered" {
		dareForWrongAnswer = &randomTrivia.DareForWrongAnswer
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

	return &model.TriviaDetail{
		ID:                 ownGame.Id.Hex(),
		Question:           randomTrivia.Question,
		Answers:            randomTrivia.Answers,
		AnsweredUsers:      answeredUsers,
		NotAnsweredUsers:   notAnsweredUsers,
		DareForWrongAnswer: dareForWrongAnswer,
		Status:             ownGame.Status,
	}, nil
}
