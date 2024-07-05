package wouldyourather

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"math"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetFilledWouldRathers(ctx context.Context, db *database.DB, page int, limit int) (*model.SavedWouldYouRatherPaginationResp, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	filledWouldYouratherColl := db.GetCollection("filledWouldYouRather")

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"userId": user.Id,
	}

	sortOptions := options.Find().SetSort(bson.M{"updatedAt": -1})
	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"updatedAt": -1})

	cursor, err := filledWouldYouratherColl.Find(ctx, filter, findOptions, sortOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.SavedWouldYouRatherPaginationResp{
				Total:                0,
				PerPage:              limit,
				CurrentPage:          page,
				TotalPages:           0,
				SavedWouldYouRathers: []*model.SavedWouldYouRatherData{},
			}, nil
		}
		return nil, gqlerror.Errorf("Failed to fetch filled wouldYouRathers")
	}
	defer cursor.Close(ctx)

	var filledWouldYouRathers []entity.FilledWouldYouRatherEntity
	var wouldYouRatherIds []primitive.ObjectID

	for cursor.Next(ctx) {
		var filledWouldYouRather entity.FilledWouldYouRatherEntity
		err := cursor.Decode(&filledWouldYouRather)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode filled wouldYouRather")
		}
		filledWouldYouRathers = append(filledWouldYouRathers, filledWouldYouRather)
		wouldYouRatherIds = append(wouldYouRatherIds, filledWouldYouRather.WouldYouRatherId)
	}

	if err := cursor.Err(); err != nil {
		return nil, gqlerror.Errorf("Cursor error: " + err.Error())
	}

	wouldYouRatherFilter := bson.M{"_id": bson.M{"$in": wouldYouRatherIds}}

	wouldYouRatherColl := db.GetCollection("wouldYouRather")
	wouldYouRatherCursor, err := wouldYouRatherColl.Find(ctx, wouldYouRatherFilter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch wouldYouRather")
	}
	defer wouldYouRatherCursor.Close(ctx)

	wouldYouRatherMap := make(map[primitive.ObjectID]entity.WouldYouRatherEntity)
	for wouldYouRatherCursor.Next(ctx) {
		var wouldYouRather entity.WouldYouRatherEntity
		err := wouldYouRatherCursor.Decode(&wouldYouRather)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode wouldYouRathers")
		}
		wouldYouRatherMap[wouldYouRather.Id] = wouldYouRather
	}

	if err := wouldYouRatherCursor.Err(); err != nil {
		return nil, gqlerror.Errorf("Cursor error while fetching wouldYouRathers: " + err.Error())
	}

	var savedWouldYouRathers []*model.SavedWouldYouRatherData
	for _, filledWouldYouRather := range filledWouldYouRathers {
		wouldYouRather, found := wouldYouRatherMap[filledWouldYouRather.WouldYouRatherId]
		if !found {
			return nil, gqlerror.Errorf("wouldYouRather not found")
		}

		savedWouldYouRatherRes := &model.SavedWouldYouRatherData{
			ID:             filledWouldYouRather.Id.Hex(),
			WouldYouRather: wouldYouRather.Question,
			Answer:         filledWouldYouRather.Answer,
		}

		savedWouldYouRathers = append(savedWouldYouRathers, savedWouldYouRatherRes)
	}

	totalCount, err := filledWouldYouratherColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count filled WouldYouRathers")
	}

	response := model.SavedWouldYouRatherPaginationResp{
		Total:                int(totalCount),
		PerPage:              limit,
		CurrentPage:          page,
		TotalPages:           int(math.Ceil(float64(totalCount) / float64(limit))),
		SavedWouldYouRathers: savedWouldYouRathers,
	}

	return &response, nil
}
