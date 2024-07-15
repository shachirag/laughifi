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

	categoryObjID, err := primitive.ObjectIDFromHex(categoryId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid category Id")
	}

	templateFilter := bson.M{"category.id": categoryObjID}
	templateCount, err := db.GetCollection("template").CountDocuments(ctx, templateFilter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to check associated templates: %s", err.Error())
	}

	if templateCount > 0 {
		return nil, gqlerror.Errorf("Cannot delete category. Templates are associated with this category.")
	}

	ownGameFilter := bson.M{"category.id": categoryObjID}
	ownGameCount, err := db.GetCollection("ownGame").CountDocuments(ctx, ownGameFilter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to check associated own game: %s", err.Error())
	}

	if ownGameCount > 0 {
		return nil, gqlerror.Errorf("Cannot delete category. game are associated with this category.")
	}

	filter := bson.M{
		"_id": categoryObjID,
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
