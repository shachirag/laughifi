package wouldyourather

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

func GetAllWouldYouRathers(ctx context.Context, db *database.DB, page int, limit int, search string) (*model.WouldYouRathersPaginationResponse, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	wouldYouRatherColl := db.GetCollection("wouldYouRather")

	filter := bson.M{
		"isDeleted": false,
	}

	if search != "blank" {
		filter["question"] = primitive.Regex{Pattern: search, Options: "i"}
	}

	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"updatedAt": -1})

	cursor, err := wouldYouRatherColl.Find(ctx, filter, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.WouldYouRathersPaginationResponse{
				Total:           0,
				PerPage:         limit,
				CurrentPage:     page,
				TotalPages:      0,
				WouldYouRathers: []*model.WouldYouRathers{},
			}, nil
		}
		return nil, gqlerror.Errorf("Failed to fetch filled wouldYouRathers")
	}
	defer cursor.Close(ctx)

	var wouldYouRathers []*model.WouldYouRathers
	for cursor.Next(ctx) {
		var wouldYouRather entity.WouldYouRatherEntity
		err := cursor.Decode(&wouldYouRather)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode wouldYouRathers")
		}

		wouldYouRatherRes := model.WouldYouRathers{
			ID:       wouldYouRather.Id.Hex(),
			Question: wouldYouRather.Question,
		}

		wouldYouRathers = append(wouldYouRathers, &wouldYouRatherRes)
	}

	totalCount, err := wouldYouRatherColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count filled wouldYouRatherRes")
	}

	response := model.WouldYouRathersPaginationResponse{
		Total:           int(totalCount),
		PerPage:         limit,
		CurrentPage:     page,
		TotalPages:      int(math.Ceil(float64(totalCount) / float64(limit))),
		WouldYouRathers: wouldYouRathers,
	}

	return &response, nil
}
