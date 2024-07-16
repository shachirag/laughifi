package trivia

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func OwnGame(ctx context.Context, db *database.DB, data model.OwnGameRequestInput) (*model.Response, error) {
	var (
		ownGameColl = db.GetCollection("ownGame")
	)

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	categoryObjID, err := primitive.ObjectIDFromHex(data.CategoryID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid category ID")
	}

	var category entity.CategoryEntity
	err = db.GetCollection("category").FindOne(ctx, bson.M{"_id": categoryObjID}).Decode(&category)
	if err != nil {
		return nil, gqlerror.Errorf("Internal server error while fetching the category: " + err.Error())
	}

	friendIdStrs := strings.Split(data.FriendIds, ",")
	var friendObjIDs []primitive.ObjectID
	for _, friendIdStr := range friendIdStrs {
		friendObjID, err := primitive.ObjectIDFromHex(friendIdStr)
		if err != nil {
			return nil, gqlerror.Errorf("invalid friend ID")
		}
		friendObjIDs = append(friendObjIDs, friendObjID)
	}

	friendDetails, err := fetchFriendDetails(ctx, db, friendObjIDs)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch friend details: %v", err)
	}
	var friends []entity.Friend
	for _, friendObjID := range friendObjIDs {
		friend := entity.Friend{
			Id:           friendObjID,
			Name:         friendDetails[friendObjID.Hex()].Name,
			Image:        friendDetails[friendObjID.Hex()].Image,
			PlayedStatus: "pending",
		}
		friends = append(friends, friend)
	}

	userAsFriend := entity.Friend{
		Id:           user.Id,
		Name:         user.Name,
		Image:        user.Image,
		PlayedStatus: "approved",
	}
	friends = append(friends, userAsFriend)

	id := primitive.NewObjectID()
	ownGame := entity.OwnGameEntity{
		Id:      id,
		Friends: friends,
		User: entity.User{
			Id:    user.Id,
			Name:  user.Name,
			Image: user.Image,
		},
		GameName: data.GameName,
		Category: entity.Category{
			Id:   categoryObjID,
			Name: category.Name,
		},
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err = ownGameColl.InsertOne(ctx, ownGame)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to add own game")
	}

	return &model.Response{
		Message: "Owned Game successfully shared",
	}, nil
}

func fetchFriendDetails(ctx context.Context, db *database.DB, friendObjIDs []primitive.ObjectID) (map[string]entity.CustomerEntity, error) {
	filter := bson.M{"_id": bson.M{"$in": friendObjIDs}}
	cursor, err := db.GetCollection("customer").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	friendDetails := make(map[string]entity.CustomerEntity)
	for cursor.Next(ctx) {
		var friend entity.CustomerEntity
		if err := cursor.Decode(&friend); err != nil {
			return nil, err
		}
		friendDetails[friend.Id.Hex()] = friend
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return friendDetails, nil
}
