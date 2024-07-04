package templates

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"math"
	"math/rand"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetTemplates(ctx context.Context, db *database.DB, page int, limit int, topic string) (*model.TemplatePaginationResponse, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	templateColl := db.GetCollection("template")

	filter := bson.M{}
	if topic != "" && topic != "random" {
		filter["topic"] = topic
	}

	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))

	cursor, err := templateColl.Find(ctx, filter, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Template not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch templates")
	}
	defer cursor.Close(ctx)

	var templates []*model.Template
	for cursor.Next(ctx) {
		var template entity.TemplateEntity
		err := cursor.Decode(&template)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode templates")
		}

		templateRes := model.Template{
			ID:       template.Id.Hex(),
			Title:    template.Title,
			Category: template.Topic,
			Template: template.Template,
		}

		templates = append(templates, &templateRes)
	}

	if topic == "random" {
		rand.Shuffle(len(templates), func(i, j int) { templates[i], templates[j] = templates[j], templates[i] })
	}

	totalCount, err := templateColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count templates")
	}

	response := model.TemplatePaginationResponse{
		Total:       int(totalCount),
		PerPage:     limit,
		CurrentPage: page,
		TotalPages:  int(math.Ceil(float64(totalCount) / float64(limit))),
		Templates:   templates,
	}

	return &response, nil
}
