package templates

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"regexp"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddTemplate(ctx context.Context, db *database.DB, data model.AdminTemplateRequestInput) (*model.Response, error) {
	var (
		templatesColl = db.GetCollection("template")
		template      entity.TemplatesEntity
	)

	categoryObjID, err := primitive.ObjectIDFromHex(data.CategoryID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid category Id")
	}

	var category entity.CategoryEntity
	err = db.GetCollection("category").FindOne(ctx, bson.M{"_id": categoryObjID}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("category not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch category")
	}

	replaceBlanks := func(template string) string {
		re := regexp.MustCompile(`\bBLANK\b`)
		counter := 1
		return re.ReplaceAllStringFunc(template, func(_ string) string {
			placeholder := fmt.Sprintf("{{BLANK%d}}", counter)
			counter++
			return placeholder
		})
	}

	processedTemplate := replaceBlanks(data.Template)

	id := primitive.NewObjectID()

	template = entity.TemplatesEntity{
		Id: id,
		Category: entity.Category{
			Id:   categoryObjID,
			Name: category.Name,
		},
		Template:  processedTemplate,
		Title:     data.Title,
		IsDeleted: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err = templatesColl.InsertOne(ctx, template)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to insert template")
	}

	return &model.Response{
		Message: "Template Saved Successfully",
	}, nil
}
