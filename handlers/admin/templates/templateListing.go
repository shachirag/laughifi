package templates

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"math"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetTemplates(ctx context.Context, db *database.DB, page int, limit int, search string) (*model.AdminTemplatePaginationResponse, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	templatesColl := db.GetCollection("templates")

	filter := bson.M{
		"isDeleted": false,
	}

	if search != "blank" {
		filter["template"] = primitive.Regex{Pattern: search, Options: "i"}
	}

	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"updatedAt": -1})

	cursor, err := templatesColl.Find(ctx, filter, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.AdminTemplatePaginationResponse{
				Total:          0,
				PerPage:        limit,
				CurrentPage:    page,
				TotalPages:     0,
				AdminTemplates: []*model.AdminTemplates{},
			}, nil
		}
		return nil, gqlerror.Errorf("Failed to fetch filled templates")
	}
	defer cursor.Close(ctx)

	var templates []*model.AdminTemplates
	for cursor.Next(ctx) {
		var template entity.TemplatesEntity
		err := cursor.Decode(&template)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode templates")
		}

		templatesRes := model.AdminTemplates{
			ID:       template.Id.Hex(),
			Category: template.Category.Name,
			Template: template.Template,
		}

		templates = append(templates, &templatesRes)
	}

	totalCount, err := templatesColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count filled templates")
	}

	response := model.AdminTemplatePaginationResponse{
		Total:          int(totalCount),
		PerPage:        limit,
		CurrentPage:    page,
		TotalPages:     int(math.Ceil(float64(totalCount) / float64(limit))),
		AdminTemplates: templates,
	}

	return &response, nil
}

func FetchCategoryDetails(ctx context.Context, db *database.DB, categoryIDs []primitive.ObjectID) (map[primitive.ObjectID]entity.CategoryEntity, error) {
	categoryDetails := make(map[primitive.ObjectID]entity.CategoryEntity)
	if len(categoryIDs) > 0 {
		customerFilter := bson.M{"_id": bson.M{"$in": categoryIDs}}
		categoryCur, err := db.GetCollection("category").Find(ctx, customerFilter)
		if err != nil {
			return nil, err
		}
		defer categoryCur.Close(ctx)

		for categoryCur.Next(ctx) {
			var category entity.CategoryEntity
			err := categoryCur.Decode(&category)
			if err != nil {
				return nil, err
			}
			categoryDetails[category.Id] = category
		}
	}
	return categoryDetails, nil
}
