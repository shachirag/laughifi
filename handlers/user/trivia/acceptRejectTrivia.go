package trivia

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"
	"laughifi/utils"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AcceptRejectTrivia(ctx context.Context, db *database.DB, id string, status string) (*model.Response, error) {

	ownGameObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid own game Id")
	}

	userData, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": ownGameObjID,
		"friends": bson.M{
			"$elemMatch": bson.M{
				"id": userData.Id,
			},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"friends.$.playedStatus": status,
		},
	}

	updateRes, err := db.GetCollection("ownGame").UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update ownGame")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("No ownGame found")
	}

	return &model.Response{
		Message: "Deleted Successfully",
	}, nil
}
