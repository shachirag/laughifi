package trivia

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"math"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAnsweredTrivia(ctx context.Context, db *database.DB, page int, limit int) (*model.AnsweredTriviaPaginationResponse, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	userData, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	ownGameColl := db.GetCollection("ownGame")

	filter := bson.M{
		"friends": bson.M{
			"$elemMatch": bson.M{
				"id":           userData.Id,
				"playedStatus": "accepted",
				"role":         "player",
			},
		},
	}

	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"updatedAt": -1})

	cursor, err := ownGameColl.Find(ctx, filter, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.AnsweredTriviaPaginationResponse{
				Total:       0,
				PerPage:     limit,
				CurrentPage: page,
				TotalPages:  0,
				Trivias:     []*model.AnsweredTrivia{},
			}, nil
		}
		return nil, gqlerror.Errorf("Failed to fetch filled trivia")
	}
	defer cursor.Close(ctx)

	var trivias []*model.AnsweredTrivia
	for cursor.Next(ctx) {
		var ownGame entity.OwnGameEntity
		err := cursor.Decode(&ownGame)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode trivia")
		}

		categoriesRes := model.AnsweredTrivia{
			ID:           ownGame.Id.Hex(),
			CategoryName: ownGame.Category.Name,
			GameName:     ownGame.GameName,
			Status:       ownGame.Status,
			Role:         "player",
		}

		trivias = append(trivias, &categoriesRes)
	}

	totalCount, err := ownGameColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count filled trivias")
	}

	response := model.AnsweredTriviaPaginationResponse{
		Total:       int(totalCount),
		PerPage:     limit,
		CurrentPage: page,
		TotalPages:  int(math.Ceil(float64(totalCount) / float64(limit))),
		Trivias:     trivias,
	}

	return &response, nil
}

// func GetAnsweredTrivia(ctx context.Context, db *database.DB) ([]*model.AnsweredTrivia, error) {
// 	var (
// 		ownGameColl = db.GetCollection("ownGame")
// 	)

// 	userData, err := utils.ExtractUserFromContext(ctx, db)
// 	if err != nil {
// 		return nil, err
// 	}

// 	filter := bson.M{
// 		"friends": bson.M{
// 			"$elemMatch": bson.M{
// 				"id":           userData.Id,
// 				"playedStatus": "accepted",
// 				"role":         "player",
// 			},
// 		},
// 	}

// 	sortOptions := options.Find().SetSort(bson.M{"updatedAt": -1})

// 	cursor, err := ownGameColl.Find(ctx, filter, sortOptions)
// 	if err != nil {
// 		return nil, gqlerror.Errorf("Failed to fetch categories")
// 	}
// 	defer cursor.Close(ctx)

// 	var triviaData []*model.AnsweredTrivia
// 	for cursor.Next(ctx) {
// 		var ownGame entity.OwnGameEntity
// 		if err := cursor.Decode(&ownGame); err != nil {
// 			return nil, gqlerror.Errorf("Failed to deocde ownGames")
// 		}

// 		triviaData = append(triviaData, &model.AnsweredTrivia{
// 			ID:           ownGame.Id.Hex(),
// 			CategoryName: ownGame.Category.Name,
// 			GameName:     ownGame.GameName,
// 			Status:       ownGame.Status,
// 			Role:         "player",
// 		})

// 	}

// 	if len(triviaData) == 0 {
// 		return []*model.AnsweredTrivia{}, nil
// 	}

// 	return triviaData, nil

// }
