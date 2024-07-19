package templates

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

func GetAllTemplates(ctx context.Context, db *database.DB, page int, limit int, search string) (*model.AdminTemplatePaginationResponse, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	templatesColl := db.GetCollection("template")

	filter := bson.M{
		"isDeleted": false,
	}

	if search != "" {
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

		displayTemplate := replacePlaceholdersWithBlank(template.Template)

		templatesRes := model.AdminTemplates{
			ID:       template.Id.Hex(),
			Category: template.Category.Name,
			Template: displayTemplate,
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

func replacePlaceholdersWithBlank(template string) string {
	re := regexp.MustCompile(`{{BLANK\d+}}`)
	return re.ReplaceAllString(template, "BLANK")
}
