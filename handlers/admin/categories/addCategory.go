package category

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddCatgeory(ctx context.Context, db *database.DB, data model.CategoryRequestInput) (*model.Response, error) {
	var (
		categoryColl = db.GetCollection("category")
		category     entity.CategoryEntity
	)

	id := primitive.NewObjectID()

	category = entity.CategoryEntity{
		Id:        id,
		Name:      data.Category,
		IsDeleted: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err := categoryColl.InsertOne(ctx, category)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to insert category")
	}

	return &model.Response{
		Message: "Category Saved Successfully",
	}, nil
}
