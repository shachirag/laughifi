package friends

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"math"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetFriends(ctx context.Context, db *database.DB, page int, limit int, search string) (*model.FriendPaginationResponse, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	customerColl := db.GetCollection("customer")

	friendIDs, err := fetchFriendIDs(ctx, db)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch friend IDs: %v", err)
	}

	if len(friendIDs) == 0 {
		return &model.FriendPaginationResponse{
			Total:       0,
			PerPage:     limit,
			CurrentPage: page,
			TotalPages:  0,
			Friends:     []*model.Friend{},
		}, nil
	}

	filter := bson.M{
		"_id": bson.M{
			"$in": friendIDs,
		},
	}

	if search != "blank" {
		filter["name"] = primitive.Regex{Pattern: search, Options: "i"}
	}

	sortOptions := options.Find().SetSort(bson.M{"updatedAt": -1})
	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"updatedAt": -1})

	cursor, err := customerColl.Find(ctx, filter, findOptions, sortOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Customer not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch customers")
	}
	defer cursor.Close(ctx)

	var friends []*model.Friend
	for cursor.Next(ctx) {
		var customer entity.CustomerEntity
		err := cursor.Decode(&customer)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode customer")
		}

		customerRes := &model.Friend{
			ID:    customer.Id.Hex(),
			Name:  customer.Name,
			Image: customer.Image,
		}

		friends = append(friends, customerRes)
	}

	totalCount, err := customerColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count customers")
	}

	response := model.FriendPaginationResponse{
		Total:       int(totalCount),
		PerPage:     limit,
		CurrentPage: page,
		TotalPages:  int(math.Ceil(float64(totalCount) / float64(limit))),
		Friends:     friends,
	}

	return &response, nil
}

func fetchFriendIDs(ctx context.Context, db *database.DB) ([]primitive.ObjectID, error) {

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"userId": user.Id,
		"friendsList": bson.M{
			"$elemMatch": bson.M{
				"status": "friend-request-accepted",
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
		if err := cursor.Decode(&friendList); err != nil {
			return nil, err
		}
		friendLists = append(friendLists, friendList)
	}

	var friendIDs []primitive.ObjectID
	for _, res := range friendLists {
		for _, friend := range res.FriendsList {
			if friend.Status == "friend-request-accepted" {
				friendIDs = append(friendIDs, friend.Id)
			}
		}
	}

	return friendIDs, nil
}
