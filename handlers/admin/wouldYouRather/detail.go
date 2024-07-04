package wouldyourather

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

func GetWouldYouRatherData(ctx context.Context, db *database.DB, wouldYouRatherId string) (*model.WouldYouRathersDetail, error) {
	var wouldYouRather entity.WouldYouRatherEntity

	wouldYouRatherObjID, err := primitive.ObjectIDFromHex(wouldYouRatherId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid wouldYouRather Id")
	}

	wouldYouRatherColl := db.GetCollection("wouldYouRather")

	err = wouldYouRatherColl.FindOne(ctx, bson.M{"_id": wouldYouRatherObjID}).Decode(&wouldYouRather)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("WouldYouRather not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch wouldYouRather")
	}

	return &model.WouldYouRathersDetail{
		ID:       wouldYouRather.Id.Hex(),
		Question: wouldYouRather.Question,
		Options:  wouldYouRather.Options,
	}, nil
}
