package category

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeletedCategoryData(ctx context.Context, db *database.DB, categoryId string) (*model.Response, error) {

	categoryObjIdID, err := primitive.ObjectIDFromHex(categoryId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid category Id")
	}

	filter := bson.M{
		"_id": categoryObjIdID,
	}

	update := bson.M{
		"$set": bson.M{
			"isDeleted": true,
		},
	}

	updateRes, err := db.GetCollection("category").UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update category")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("No category found")
	}

	return &model.Response{
		Message: "Catgeory deleted Successfully",
	}, nil
}
