package trivia

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeleteTrivia(ctx context.Context, db *database.DB, triviaId string) (*model.Response, error) {

	triviaObjID, err := primitive.ObjectIDFromHex(triviaId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid trivia Id")
	}

	filter := bson.M{
		"_id": triviaObjID,
	}

	update := bson.M{
		"$set": bson.M{
			"isDeleted": true,
		},
	}

	updateRes, err := db.GetCollection("trivia").UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update trivia")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("No trivia found")
	}

	return &model.Response{
		Message: "Trivia deleted Successfully",
	}, nil
}
