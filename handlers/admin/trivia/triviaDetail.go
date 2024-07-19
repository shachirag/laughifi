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

func GetAdminTrivia(ctx context.Context, db *database.DB, triviaId string) (*model.TriviaData, error) {
	var trivia entity.TriviaEntity

	triviaObjID, err := primitive.ObjectIDFromHex(triviaId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid trivia Id")
	}

	triviaColl := db.GetCollection("trivia")

	err = triviaColl.FindOne(ctx, bson.M{"_id": triviaObjID}).Decode(&trivia)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("trivia not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch trivia")
	}

	return &model.TriviaData{
		ID:                 trivia.Id.Hex(),
		Question:           trivia.Question,
		Answers:            trivia.Answers,
		Category:           trivia.Category.Name,
		CategoryID:         trivia.Category.Id.Hex(),
		CorrectAnswer:      trivia.CorrectAnswer,
		DareForWrongAnswer: trivia.DareForWrongAnswer,
	}, nil
}
