package friends

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"
	"laughifi/utils"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CancelFriendRequest(ctx context.Context, db *database.DB, data model.SendFriendRequestInput) (*model.Response, error) {
	var (
		friendsListColl = db.GetCollection("friendsList")
	)

	userObjID, err := primitive.ObjectIDFromHex(data.UserID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid user Id")
	}

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"userId": user.Id,
		"friendsList": bson.M{
			"$elemMatch": bson.M{
				"id":     userObjID,
				"status": "friend-request-pending",
			},
		},
	}

	update := bson.M{
		"$pull": bson.M{
			"friendsList": bson.M{
				"id":     userObjID,
				"status": "friend-request-pending",
			},
		},
	}

	session, err := db.GetMongoClient().StartSession()
	if err != nil {
		return nil, gqlerror.Errorf("Failed to start session")
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		result, err := friendsListColl.UpdateOne(sessCtx, filter, update)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to cancel friend request: %v", err)
		}

		if result.MatchedCount == 0 {
			return nil, gqlerror.Errorf("No matching friend request found to cancel")
		}

		secondFilter := bson.M{
			"userId": userObjID,
			"friendsList": bson.M{
				"$elemMatch": bson.M{
					"id":     user.Id,
					"status": "friend-request-pending",
				},
			},
		}

		secondUpdate := bson.M{
			"$pull": bson.M{
				"friendsList": bson.M{
					"id":     user.Id,
					"status": "friend-request-pending",
				},
			},
		}

		secondResult, err := friendsListColl.UpdateOne(sessCtx, secondFilter, secondUpdate)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to cancel friend request: %v", err)
		}

		if secondResult.MatchedCount == 0 {
			return nil, gqlerror.Errorf("No matching friend request found to cancel for the second user")
		}

		return nil, nil
	}

	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		return nil, gqlerror.Errorf("Transaction failed: %v", err)
	}

	return &model.Response{
		Message: "success",
	}, nil
}
