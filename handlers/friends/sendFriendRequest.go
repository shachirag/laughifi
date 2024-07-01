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

func SendFriendRequest(ctx context.Context, db *database.DB, data model.SendFriendRequestInput) (*model.Response, error) {
	var (
		friendsListColl = db.GetCollection("friendsList")
	)

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
		}

		for _, memberId := range data.UserIds {
			memberObjID, err := primitive.ObjectIDFromHex(memberId)
			if err != nil {
				return nil, gqlerror.Errorf("Invalid user ID: %v", err)
			}

			newFriendList.FriendsList = append(newFriendList.FriendsList, entity.FriendsList{
				Id:     memberObjID,
				Status: "friend-request-pending",
			})
		}

		_, err = friendsListColl.InsertOne(ctx, newFriendList)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to send friend request: %v", err)
		}
	} else if err != nil {
		return nil, gqlerror.Errorf("Failed to check existing friend requests: %v", err)
	} else {

		var duplicateUserIDs []string

		for _, memberId := range data.UserIds {
			memberObjID, err := primitive.ObjectIDFromHex(memberId)
			if err != nil {
				return nil, gqlerror.Errorf("Invalid user ID: %v", err)
			}

			for _, existingFriend := range existingFriendList.FriendsList {
				if existingFriend.Id == memberObjID && existingFriend.Status == "friend-request-pending" {
					duplicateUserIDs = append(duplicateUserIDs, memberId)
					break
				}
			}
		}

		if len(duplicateUserIDs) > 0 {
			return nil, gqlerror.Errorf("Friend request already sent")
		}

		var updates []entity.FriendsList

		for _, memberId := range data.UserIds {
			memberObjID, err := primitive.ObjectIDFromHex(memberId)
			if err != nil {
				return nil, gqlerror.Errorf("Invalid user ID: %v", err)
			}

			exists := false
			for _, existingFriend := range existingFriendList.FriendsList {
				if existingFriend.Id == memberObjID {
					exists = true
					break
				}
			}

			if !exists {
				updates = append(updates, entity.FriendsList{
					Id:     memberObjID,
					Status: "friend-request-pending",
				})
			}
		}

		if len(updates) > 0 {
			update := bson.M{
				"$push": bson.M{
					"friendsList": bson.M{
						"$each": updates,
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
	}

	return &model.Response{
		Message: "Friend Request Sent Successfully",
	}, nil
}
