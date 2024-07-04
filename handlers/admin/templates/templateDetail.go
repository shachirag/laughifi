package templates

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

func GetTemplateData(ctx context.Context, db *database.DB, templateId string) (*model.AdminTemplate, error) {
	var template entity.TemplatesEntity

	templateObjID, err := primitive.ObjectIDFromHex(templateId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid template Id")
	}

	templateColl := db.GetCollection("templates")

	err = templateColl.FindOne(ctx, bson.M{"_id": templateObjID}).Decode(&template)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("template not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch template")
	}

	var category entity.CategoryEntity
	err = db.GetCollection("category").FindOne(ctx, bson.M{"_id": template.Category.Id}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("category not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch category")
	}

	return &model.AdminTemplate{
		ID:       template.Id.Hex(),
		Category: category.Name,
		Title:    template.Title,
		Template: template.Template,
	}, nil
}
