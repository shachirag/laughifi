package friends

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"
	"laughifi/utils"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	_, err = friendsListColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to cancel friend request: %v", err)
	}

	return &model.Response{
		Message: "success",
	}, nil
}
