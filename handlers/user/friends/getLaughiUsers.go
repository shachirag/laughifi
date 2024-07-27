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

func GetLaughifiCustomers(ctx context.Context, db *database.DB, page int, limit int, search string) (*model.LaughifiUserPaginationResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	customerColl := db.GetCollection("customer")

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	friendIDs, err := fetchLaughifiUsersIDs(ctx, db)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch friend IDs: %v", err)
	}

	excludeIDs := append(friendIDs, user.Id)

	filter := bson.M{
		"_id": bson.M{
			"$nin": excludeIDs,
		},
		"isDeleted": false,
	}

	if search != "blank" {
		filter["name"] = primitive.Regex{Pattern: search, Options: "i"}
	}

	sortOptions := options.Find().SetSort(bson.M{"updatedAt": -1})
	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))

	cursor, err := customerColl.Find(ctx, filter, findOptions, sortOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Customer not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch customers: %v", err)
	}
	defer cursor.Close(ctx)

	var customers []entity.CustomerEntity
	customerIDs := make([]primitive.ObjectID, 0)
	for cursor.Next(ctx) {
		var customer entity.CustomerEntity
		if err := cursor.Decode(&customer); err != nil {
			return nil, gqlerror.Errorf("Failed to decode customer: %v", err)
		}
		customers = append(customers, customer)
		customerIDs = append(customerIDs, customer.Id)
	}

	// Determine pending friend requests from the current user to customers
	getRequestFilter := bson.M{
		"userId": user.Id,
		"friendsList": bson.M{
			"$elemMatch": bson.M{
				"id": bson.M{
					"$in": customerIDs,
				},
				"status": "friend-request-pending",
			},
		},
	}

	cursor, err = db.GetCollection("friendsList").Find(ctx, getRequestFilter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch pending friend requests: %v", err)
	}
	defer cursor.Close(ctx)

	friendRequestStatus := make(map[primitive.ObjectID]bool)
	for cursor.Next(ctx) {
		var friendsList entity.FriendListEntity
		if err := cursor.Decode(&friendsList); err != nil {
			return nil, gqlerror.Errorf("Failed to decode friends list: %v", err)
		}
		for _, friend := range friendsList.FriendsList {
			if friend.Status == "friend-request-pending" {
				friendRequestStatus[friend.Id] = true
			}
		}
	}

	// Determine the current user's status in the context of customers
	statusFilter := bson.M{
		"userId": bson.M{
			"$in": customerIDs,
		},
		"friendsList": bson.M{
			"$elemMatch": bson.M{
				"id": user.Id,
			},
		},
	}

	cursor, err = db.GetCollection("friendsList").Find(ctx, statusFilter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch current user status: %v", err)
	}
	defer cursor.Close(ctx)

	customerStatus := make(map[primitive.ObjectID]string)
	for cursor.Next(ctx) {
		var friendsList entity.FriendListEntity
		if err := cursor.Decode(&friendsList); err != nil {
			return nil, gqlerror.Errorf("Failed to decode friends list: %v", err)
		}
		for _, friend := range friendsList.FriendsList {
			if friend.Id == user.Id {
				customerStatus[friendsList.UserId] = friend.Status
			}
		}
	}

	var templates []*model.FriendLaughifiUsers
	for _, customer := range customers {
		getRequest := friendRequestStatus[customer.Id]

		customerRes := &model.FriendLaughifiUsers{
			ID:         customer.Id.Hex(),
			Name:       customer.Name,
			Image:      customer.Image,
			Status:     customerStatus[customer.Id] == "friend-request-pending",
			GetRequest: getRequest,
		}
		templates = append(templates, customerRes)
	}

	totalCount, err := customerColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count customers: %v", err)
	}

	response := model.LaughifiUserPaginationResponse{
		Total:       int(totalCount),
		PerPage:     limit,
		CurrentPage: page,
		TotalPages:  int(math.Ceil(float64(totalCount) / float64(limit))),
		Users:       templates,
	}

	return &response, nil
}

func fetchLaughifiUsersIDs(ctx context.Context, db *database.DB) ([]primitive.ObjectID, error) {
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
		return nil, gqlerror.Errorf("Failed to find friends list: %v", err)
	}
	defer cursor.Close(ctx)

	var friendLists []entity.FriendListEntity
	for cursor.Next(ctx) {
		var friendList entity.FriendListEntity
		if err := cursor.Decode(&friendList); err != nil {
			return nil, gqlerror.Errorf("Failed to decode friends list: %v", err)
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
