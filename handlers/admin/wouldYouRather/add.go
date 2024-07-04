package wouldyourather

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddWouldYouRather(ctx context.Context, db *database.DB, data model.WouldYouRatherRequestInput) (*model.Response, error) {
	var (
		wouldYouRatherColl = db.GetCollection("wouldYouRather")
		wouldYouRather     entity.WouldYouRatherEntity
	)

	id := primitive.NewObjectID()

	wouldYouRather = entity.WouldYouRatherEntity{
		Id:        id,
		Question:  data.Question,
		Options:   data.Options,
		IsDeleted: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err := wouldYouRatherColl.InsertOne(ctx, wouldYouRather)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to insert wouldYouRather")
	}

	return &model.Response{
		Message: "WouldYouRather Saved Successfully",
	}, nil
}
