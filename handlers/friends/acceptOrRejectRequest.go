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

func AcceptRejectRequest(ctx context.Context, db *database.DB, userId string, data model.UpdateStatusRequestInput) (*model.Response, error) {

	var (
		friendsList = db.GetCollection("friendsList")
	)

	userObjIdID, err := primitive.ObjectIDFromHex(userId)
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
				"id":     userObjIdID,
				"status": bson.M{"$in": []string{"friend-request-pending", "friend-request-accepted"}},
			},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"friendsList.$.status": data.Status,
		},
	}

	_, err = friendsList.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update status")
	}

	var actionMessage string
	if data.Status == "friend-request-accepted" {
		actionMessage = "Friend request accepted"
	} else if data.Status == "friend-request-rejected" {
		actionMessage = "Friend request rejected"
	} else if data.Status == "friend-removed" {
		actionMessage = "Friend removed."
	}

	return &model.Response{
		Message: actionMessage,
	}, nil

}
