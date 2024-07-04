package wouldyourather

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeleteWouldYouRather(ctx context.Context, db *database.DB, wouldYouRatherId string) (*model.Response, error) {

	wouldYouRatherObjID, err := primitive.ObjectIDFromHex(wouldYouRatherId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid wouldYouRather Id")
	}

	filter := bson.M{
		"_id": wouldYouRatherObjID,
	}

	update := bson.M{
		"$set": bson.M{
			"isDeleted": true,
		},
	}

	updateRes, err := db.GetCollection("wouldYouRather").UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update wouldYouRather")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("No wouldYouRather found")
	}

	return &model.Response{
		Message: "WouldYouRather deleted Successfully",
	}, nil
}
