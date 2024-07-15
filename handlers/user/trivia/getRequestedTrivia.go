package trivia

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetRequestedTrivia(ctx context.Context, db *database.DB) ([]*model.RequestedTrivia, error) {
	var (
		ownGameColl = db.GetCollection("ownGame")
	)

	userData, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"friends": bson.M{
			"$elemMatch": bson.M{
				"id":     userData.Id,
				"playedStatus": "pending",
			},
		},
	}

	sortOptions := options.Find().SetSort(bson.M{"updatedAt": -1})

	cursor, err := ownGameColl.Find(ctx, filter, sortOptions)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch categories")
	}
	defer cursor.Close(ctx)

	var triviaData []*model.RequestedTrivia
	for cursor.Next(ctx) {
		var ownGame entity.OwnGameEntity
		if err := cursor.Decode(&ownGame); err != nil {
			return nil, gqlerror.Errorf("Failed to deocde ownGames")
		}

		triviaData = append(triviaData, &model.RequestedTrivia{
			ID:     ownGame.Id.Hex(),
			Status: ownGame.Status,
		})

	}

	if len(triviaData) == 0 {
		return []*model.RequestedTrivia{}, nil
	}

	return triviaData, nil

}
