package templates

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeletedTemplateData(ctx context.Context, db *database.DB, templateId string) (*model.Response, error) {

	templateObjIdID, err := primitive.ObjectIDFromHex(templateId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid template Id")
	}

	filter := bson.M{
		"_id": templateObjIdID,
	}

	update := bson.M{
		"$set": bson.M{
			"isDeleted": true,
		},
	}

	updateRes, err := db.GetCollection("template").UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update template")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("No template found")
	}

	return &model.Response{
		Message: "Template deleted Successfully",
	}, nil
}
