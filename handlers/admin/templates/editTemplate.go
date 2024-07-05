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

func EditTemplate(ctx context.Context, db *database.DB, templateId string, input model.EditAdminTemplateRequestInput) (*model.Response, error) {
	var (
		templatesColl = db.GetCollection("template")
	)

	templateObjID, err := primitive.ObjectIDFromHex(templateId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid template Id")
	}

	categoryObjID, err := primitive.ObjectIDFromHex(input.CategoryID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid template Id")
	}

	var category entity.CategoryEntity
	err = db.GetCollection("category").FindOne(ctx, bson.M{"_id": categoryObjID}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("category not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch category")
	}

	filter := bson.M{"_id": templateObjID}

	replaceBlanks := func(template string) string {
		re := regexp.MustCompile(`\bBLANK\b`)
		counter := 1
		return re.ReplaceAllStringFunc(template, func(_ string) string {
			placeholder := fmt.Sprintf("{{BLANK%d}}", counter)
			counter++
			return placeholder
		})
	}

	processedTemplate := replaceBlanks(input.Template)

	update := bson.M{
		"$set": bson.M{
			"category.id":   categoryObjID,
			"category.name": category.Name,
			"title":         input.Title,
			"template":      processedTemplate,
			"updatedAt":     time.Now().UTC(),
		},
	}

	updateRes, err := templatesColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update template")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("No template found")
	}

	return &model.Response{
		Message: "Template Updated Successfully",
	}, nil
}
