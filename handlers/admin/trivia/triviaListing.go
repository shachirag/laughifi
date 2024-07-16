package trivia

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"math"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetTrivias(ctx context.Context, db *database.DB, page int, limit int, search string) (*model.TriviaPaginationResponse, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	categoryColl := db.GetCollection("category")

	filter := bson.M{
		"isDeleted": false,
	}

	if search != "blank" {
		filter["question"] = primitive.Regex{Pattern: search, Options: "i"}
	}

	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"updatedAt": -1})

	cursor, err := categoryColl.Find(ctx, filter, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.TriviaPaginationResponse{
				Total:       0,
				PerPage:     limit,
				CurrentPage: page,
				TotalPages:  0,
				Trivias:     []*model.AdminTriviaListing{},
			}, nil
		}
		return nil, gqlerror.Errorf("Failed to fetch filled categories")
	}
	defer cursor.Close(ctx)

	var trivias []*model.AdminTriviaListing
	for cursor.Next(ctx) {
		var trivia entity.TriviaEntity
		err := cursor.Decode(&trivia)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode trivias")
		}

		triviaRes := model.AdminTriviaListing{
			ID:       trivia.Id.Hex(),
			Category: trivia.Category.Name,
			Question: trivia.Question,
		}

		trivias = append(trivias, &triviaRes)
	}

	totalCount, err := categoryColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count filled trivias")
	}

	response := model.TriviaPaginationResponse{
		Total:       int(totalCount),
		PerPage:     limit,
		CurrentPage: page,
		TotalPages:  int(math.Ceil(float64(totalCount) / float64(limit))),
		Trivias:     trivias,
	}

	return &response, nil
}
