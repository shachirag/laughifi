package category

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetCategoryData(ctx context.Context, db *database.DB, categoryId string) (*model.Category, error) {
	var category entity.CategoryEntity

	categoryObjIdID, err := primitive.ObjectIDFromHex(categoryId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid user Id")
	}

	categoryColl := db.GetCollection("category")

	err = categoryColl.FindOne(ctx, bson.M{"_id": categoryObjIdID}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Category not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch category")
	}

	return &model.Category{
		ID:       category.Id.Hex(),
		Category: category.Name,
	}, nil
}
