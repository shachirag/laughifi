package wouldyourather

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"math"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetWouldYouRathers(ctx context.Context, db *database.DB, page int, limit int) (*model.WouldYouRatherPaginationResp, error) {

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

	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))

	cursor, err := wouldYouRatherColl.Find(ctx, filter, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("wouldYouRather not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch wouldYouRathers")
	}
	defer cursor.Close(ctx)

	var WouldYouRathers []*model.WouldYouRatherData
	for cursor.Next(ctx) {
		var WouldYouRather entity.WouldYouRatherEntity
		err := cursor.Decode(&WouldYouRather)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode wouldYouRathers")
		}

		WouldYouRatherRes := model.WouldYouRatherData{
			ID:             WouldYouRather.Id.Hex(),
			WouldYouRather: WouldYouRather.Question,
			Options:        WouldYouRather.Options,
		}

		WouldYouRathers = append(WouldYouRathers, &WouldYouRatherRes)
	}

	totalCount, err := wouldYouRatherColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count templates")
	}

	response := model.WouldYouRatherPaginationResp{
		Total:           int(totalCount),
		PerPage:         limit,
		CurrentPage:     page,
		TotalPages:      int(math.Ceil(float64(totalCount) / float64(limit))),
		WouldYouRathers: WouldYouRathers,
	}

	return &response, nil
}
