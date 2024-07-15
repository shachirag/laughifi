package friends

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SendFriendRequest(ctx context.Context, db *database.DB, data model.SendFriendRequestInput) (*model.RequestResponse, error) {
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

	var existingFriendList entity.FriendListEntity
	err = friendsListColl.FindOne(ctx, bson.M{"userId": user.Id}).Decode(&existingFriendList)

	if err == mongo.ErrNoDocuments {
		newFriendList := entity.FriendListEntity{
			Id:        primitive.NewObjectID(),
			UserId:    user.Id,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			FriendsList: []entity.FriendsList{
				{
					Id:     userObjID,
					Status: "friend-request-pending",
				},
			},
		}

		_, err = friendsListColl.InsertOne(ctx, newFriendList)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to send friend request: %v", err)
		}
	} else if err != nil {
		return nil, gqlerror.Errorf("Failed to check existing friend requests: %v", err)
	} else {
		for _, existingFriend := range existingFriendList.FriendsList {
			if existingFriend.Id.Hex() == data.UserID {
				if existingFriend.Status == "friend-request-pending" {
					return &model.RequestResponse{
						Message: "Not Shared",
					}, nil
				}
			}
		}

		update := bson.M{
			"$push": bson.M{
				"friendsList": bson.M{
					"id":     userObjID,
					"status": "friend-request-pending",
				},
			},
			"$set": bson.M{
				"updatedAt": time.Now().UTC(),
			},
		}

		_, err = friendsListColl.UpdateOne(ctx, bson.M{"userId": user.Id}, update)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to update friend request: %v", err)
		}
	}

	var targetFriendList entity.FriendListEntity
	err = friendsListColl.FindOne(ctx, bson.M{"userId": userObjID}).Decode(&targetFriendList)

	if err == mongo.ErrNoDocuments {
		newFriendList := entity.FriendListEntity{
			Id:        primitive.NewObjectID(),
			UserId:    userObjID,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			FriendsList: []entity.FriendsList{
				{
					Id:     user.Id,
					Status: "friend-request-pending",
				},
			},
		}

		_, err = friendsListColl.InsertOne(ctx, newFriendList)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to send friend request to target user: %v", err)
		}
	} else if err != nil {
		return nil, gqlerror.Errorf("Failed to check target user's friend requests: %v", err)
	} else {
		for _, existingFriend := range targetFriendList.FriendsList {
			if existingFriend.Id.Hex() == user.Id.Hex() {
				if existingFriend.Status == "friend-request-pending" {
					return &model.RequestResponse{
						Message: "Not Shared",
					}, nil
				}
			}
		}

		update := bson.M{
			"$push": bson.M{
				"friendsList": bson.M{
					"id":     user.Id,
					"status": "friend-request-pending",
				},
			},
			"$set": bson.M{
				"updatedAt": time.Now().UTC(),
			},
		}

		_, err = friendsListColl.UpdateOne(ctx, bson.M{"userId": userObjID}, update)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to update target user's friend request: %v", err)
		}
	}

	return &model.RequestResponse{
		Message: "Shared",
	}, nil
}
