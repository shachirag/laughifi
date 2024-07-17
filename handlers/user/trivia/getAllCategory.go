package trivia

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAllCategory(ctx context.Context, db *database.DB, search string) ([]*model.GetAllCategories, error) {
	var (
		categoryColl = db.GetCollection("category")
	)

	filter := bson.M{
		"isDeleted": false,
	}

	if search != "blank" {
		filter["name"] = primitive.Regex{Pattern: search, Options: "i"}
	}

	sortOptions := options.Find().SetSort(bson.M{"updatedAt": -1})

	cursor, err := categoryColl.Find(ctx, filter, sortOptions)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch categories")
	}
	defer cursor.Close(ctx)

	var categoryData []*model.GetAllCategories
	for cursor.Next(ctx) {
		var category entity.CategoryEntity
		if err := cursor.Decode(&category); err != nil {
			return nil, gqlerror.Errorf("Failed to deocde categories")
		}

		categoryData = append(categoryData, &model.GetAllCategories{
			ID:       category.Id.Hex(),
			Category: category.Name,
		})

	}

	if len(categoryData) == 0 {
		return []*model.GetAllCategories{}, nil
	}

	for i, cat := range categoryData {
		if cat.Category == "All" {
			categoryData = append([]*model.GetAllCategories{cat}, append(categoryData[:i], categoryData[i+1:]...)...)
			break
		}
	}

	return categoryData, nil

}
