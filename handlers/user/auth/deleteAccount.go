package auth

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"
	"laughifi/utils"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

	session, err := db.GetMongoClient().StartSession()
	if err != nil {
		return nil, gqlerror.Errorf("Failed to start session")
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {

		updateRes, err := customerColl.UpdateOne(sessCtx, filter, update)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to update users")
		}

		if updateRes.MatchedCount == 0 {
			return nil, gqlerror.Errorf("No users found")
		}

		friendFilter := bson.M{
			"friendsList": bson.M{
				"$elemMatch": bson.M{
					"id": userData.Id,
				},
			},
		}

		friendUpdate := bson.M{
			"$pull": bson.M{
				"friendsList": bson.M{
					"id": userData.Id,
				},
			},
		}

		_, err = db.GetCollection("friendsList").UpdateMany(sessCtx, friendFilter, friendUpdate)
		if err != nil && err != mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Failed to update friends list")
		}

		return nil, nil
	}

	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		return nil, gqlerror.Errorf("Transaction failed: %v", err)
	}

	return &model.Response{
		Message: "Account deleted successfully",
	}, nil
}
