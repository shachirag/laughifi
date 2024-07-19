package category

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"math"
	"regexp"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetCategoryData(ctx context.Context, db *database.DB, page int, limit int, categoryId string) (*model.CategoryDetailPaginationResponse, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	categoryObjID, err := primitive.ObjectIDFromHex(categoryId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid category Id")
	}

	var category entity.CategoryEntity

	categoryColl := db.GetCollection("category")

	err = categoryColl.FindOne(ctx, bson.M{"_id": categoryObjID}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("category not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch category")
	}

	templateColl := db.GetCollection("template")

	filter := bson.M{
		"isDeleted":   false,
		"category.id": category.Id,
	}

	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"updatedAt": -1})

	cursor, err := templateColl.Find(ctx, filter, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.CategoryDetailPaginationResponse{
				Total:       0,
				PerPage:     limit,
				CurrentPage: page,
				TotalPages:  0,
				Templates:   []*model.Templates{},
			}, nil
		}
		return nil, gqlerror.Errorf("Failed to fetch filled categories")
	}
	defer cursor.Close(ctx)

	var templates []*model.Templates
	for cursor.Next(ctx) {
		var template entity.TemplatesEntity
		err := cursor.Decode(&template)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode categories")
		}

		displayTemplate := replacePlaceholdersWithBlank(template.Template)

		categoriesRes := model.Templates{
			ID:       template.Id.Hex(),
			Template: displayTemplate,
		}

		templates = append(templates, &categoriesRes)
	}

	totalCount, err := templateColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count templates")
	}

	response := model.CategoryDetailPaginationResponse{
		Total:       int(totalCount),
		PerPage:     limit,
		CurrentPage: page,
		TotalPages:  int(math.Ceil(float64(totalCount) / float64(limit))),
		Category: &model.Category{
			ID:       categoryId,
			Category: category.Name,
		},
		Templates: templates,
	}

	return &response, nil
}

func replacePlaceholdersWithBlank(template string) string {
	re := regexp.MustCompile(`{{BLANK\d+}}`)
	return re.ReplaceAllString(template, "BLANK")
}
