package auth

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"
	"laughifi/utils"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
)

func DeleteAccount(ctx context.Context, db *database.DB) (*model.Response, error) {
	var (
		customerColl = db.GetCollection("customer")
	)

	userData, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": userData.Id,
	}

	update := bson.M{
		"$set": bson.M{
			"isDeleted": true,
			"updatedAt": time.Now().UTC(),
		},
	}

	updateRes, err := customerColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update users")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("No users found")
	}

	return &model.Response{
		Message: "Account deleted successfully",
	}, nil
}
