package templates

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SubscriptionConnection(ctx context.Context, db *database.DB) (*model.Response, error) {
	var (
		subscriptionConnectionColl = db.GetCollection("subscriptionConnection")
		// playWithFriendTemplate     entity.TemplatePlayWithFriendEntity
	)

	// playWithFriendTemplateObjId, err := primitive.ObjectIDFromHex(id)
	// if err != nil {
	// 	return nil, gqlerror.Errorf("invalid template Id")
	// }

	newAnswer := entity.AnswersTemplatePlayWithFriendEntity{
		Id: primitive.NewObjectID(),
		// Key:   data.Key,
		// Value: data.Value,
	}

	_, err := subscriptionConnectionColl.InsertOne(ctx, newAnswer)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to implement subscription")
	}

	return &model.Response{
		Message: "Success",
	}, nil
}
