package trivia

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

func GetRequestedTrivia(ctx context.Context, db *database.DB) ([]*model.RequestedTrivia, error) {
	var (
		ownGameColl = db.GetCollection("ownGame")
	)

	userData, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"friends": bson.M{
			"$elemMatch": bson.M{
				"id":           userData.Id,
				"playedStatus": "pending",
				"role":         "player",
			},
		},
	}

	sortOptions := options.Find().SetSort(bson.M{"updatedAt": -1})

	cursor, err := ownGameColl.Find(ctx, filter, sortOptions)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch categories")
	}
	defer cursor.Close(ctx)

	var userIds []primitive.ObjectID
	var triviaData []*model.RequestedTrivia
	ownGames := make(map[primitive.ObjectID]entity.OwnGameEntity)
	for cursor.Next(ctx) {
		var ownGame entity.OwnGameEntity
		if err := cursor.Decode(&ownGame); err != nil {
			return nil, gqlerror.Errorf("Failed to decode ownGames")
		}

		userIds = append(userIds, ownGame.User.Id)
		ownGames[ownGame.Id] = ownGame

		triviaData = append(triviaData, &model.RequestedTrivia{
			ID:     ownGame.Id.Hex(),
			Status: ownGame.Status,
		})
	}

	if len(userIds) > 0 {
		friendDetails, err := fetcheFriendDetails(ctx, db, userIds)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to fetch friend details: %v", err)
		}

		if len(triviaData) == 0 {
			return []*model.RequestedTrivia{}, nil
		}

		for i, trivia := range triviaData {
			triviaObjID, err := primitive.ObjectIDFromHex(trivia.ID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid user Id")
			}
			ownGame := ownGames[triviaObjID]
			userDetail, exists := friendDetails[ownGame.User.Id]
			if exists {
				triviaData[i].User = &model.UserInfo{
					ID:    userDetail.Id.Hex(),
					Name:  userDetail.Name,
					Image: userDetail.Image,
				}
			} else {
				triviaData[i].User = &model.UserInfo{
					ID:    "",
					Name:  "Unknown",
					Image: "",
				}
			}
		}
	}

	return triviaData, nil
}

func fetcheFriendDetails(ctx context.Context, db *database.DB, userIds []primitive.ObjectID) (map[primitive.ObjectID]*entity.CustomerEntity, error) {
	var (
		usersColl = db.GetCollection("customer")
		userMap   = make(map[primitive.ObjectID]*entity.CustomerEntity)
	)

	filter := bson.M{"_id": bson.M{"$in": userIds}}

	cursor, err := usersColl.Find(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch users")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user entity.CustomerEntity
		if err := cursor.Decode(&user); err != nil {
			return nil, gqlerror.Errorf("Failed to decode user")
		}
		userMap[user.Id] = &user
	}

	return userMap, nil
}
