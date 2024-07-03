package friends

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetFriendRequests(ctx context.Context, db *database.DB) ([]*model.Friend, error) {
	var (
		customerColl = db.GetCollection("customer")
	)

	friendIDs, err := fetchPendingRequestFriendIDs(ctx, db)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch friend IDs: %v", err)
	}

	if len(friendIDs) == 0 {
		return []*model.Friend{}, nil
	}

	filter := bson.M{}
	filter["_id"] = bson.M{
		"$in": friendIDs,
	}

	sortOptions := options.Find().SetSort(bson.M{"updatedAt": -1})

	cursor, err := customerColl.Find(ctx, filter, sortOptions)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch friend requests: %v", err)
	}
	defer cursor.Close(ctx)

	var friendRequestsData []*model.Friend
	for cursor.Next(ctx) {
		var customer entity.CustomerEntity
		if err := cursor.Decode(&customer); err != nil {
			return nil, gqlerror.Errorf("Failed to deocde friend requests")
		}

		friendRequestsData = append(friendRequestsData, &model.Friend{
			ID:    customer.Id.Hex(),
			Name:  customer.Name,
			Image: customer.Image,
		})

	}

	if len(friendRequestsData) == 0 {
		return []*model.Friend{}, nil
	}

	return friendRequestsData, nil

}

func fetchPendingRequestFriendIDs(ctx context.Context, db *database.DB) ([]primitive.ObjectID, error) {

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"friendsList": bson.M{
			"$elemMatch": bson.M{
				"id":     user.Id,
				"status": "friend-request-pending",
			},
		},
	}

	cursor, err := db.GetCollection("friendsList").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var friendLists []entity.FriendListEntity

	for cursor.Next(ctx) {
		var friendList entity.FriendListEntity
		err := cursor.Decode(&friendList)
		if err != nil {
			return nil, err
		}
		friendLists = append(friendLists, friendList)
	}

	var friendIDs []primitive.ObjectID
	for _, res := range friendLists {
		for _, friend := range res.FriendsList {
			if friend.Status == "friend-request-pending" {
				friendIDs = append(friendIDs, friend.Id)
			}
		}
	}

	return friendIDs, nil
}
