package wouldyourather

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func FilledWouldYouRather(ctx context.Context, db *database.DB, data model.FilledWouldYouRatherRequestInput) (*model.Response, error) {
	var (
		filledWouldYouRatherColl = db.GetCollection("filledWouldYouRather")
		wouldYouRather           entity.FilledWouldYouRatherEntity
	)

	wouldYouRatherObjId, err := primitive.ObjectIDFromHex(data.WouldYouRatherID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid wouldYouRather Id")
	}

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	id := primitive.NewObjectID()

	wouldYouRather = entity.FilledWouldYouRatherEntity{
		Id:               id,
		UserId:           user.Id,
		WouldYouRatherId: wouldYouRatherObjId,
		Answer:           data.Answer,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	_, err = filledWouldYouRatherColl.InsertOne(ctx, wouldYouRather)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to save wouldYouRather")
	}

	return &model.Response{
		Message: "WouldYouRather Saved Successfully",
	}, nil
}
