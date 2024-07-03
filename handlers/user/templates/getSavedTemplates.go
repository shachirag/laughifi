package templates

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"math"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetFilledTemplates(ctx context.Context, db *database.DB, page int, limit int) (*model.SavedTemplatePaginationResponse, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	filledTemplateColl := db.GetCollection("filledTemplate")

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"userId": user.Id,
	}

	sortOptions := options.Find().SetSort(bson.M{"updatedAt": -1})
	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"updatedAt": -1})

	cursor, err := filledTemplateColl.Find(ctx, filter, findOptions, sortOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.SavedTemplatePaginationResponse{
				Total:          0,
				PerPage:        limit,
				CurrentPage:    page,
				TotalPages:     0,
				SavedTemplates: []*model.SavedTemplatesData{},
			}, nil
		}
		return nil, gqlerror.Errorf("Failed to fetch filled templates")
	}
	defer cursor.Close(ctx)

	var filledTemplates []entity.FilledTemplateEntity
	var templateIds []primitive.ObjectID

	for cursor.Next(ctx) {
		var filledTemplate entity.FilledTemplateEntity
		err := cursor.Decode(&filledTemplate)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode filled template")
		}
		filledTemplates = append(filledTemplates, filledTemplate)
		templateIds = append(templateIds, filledTemplate.TemplateId)
	}

	if err := cursor.Err(); err != nil {
		return nil, gqlerror.Errorf("Cursor error: " + err.Error())
	}

	templateFilter := bson.M{"_id": bson.M{"$in": templateIds}}

	templateColl := db.GetCollection("template")
	templateCursor, err := templateColl.Find(ctx, templateFilter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch templates")
	}
	defer templateCursor.Close(ctx)

	templateMap := make(map[primitive.ObjectID]entity.TemplateEntity)
	for templateCursor.Next(ctx) {
		var template entity.TemplateEntity
		err := templateCursor.Decode(&template)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode template")
		}
		templateMap[template.Id] = template
	}

	if err := templateCursor.Err(); err != nil {
		return nil, gqlerror.Errorf("Cursor error while fetching templates: " + err.Error())
	}

	var savedTemplates []*model.SavedTemplatesData
	for _, filledTemplate := range filledTemplates {
		template, found := templateMap[filledTemplate.TemplateId]
		if !found {
			return nil, gqlerror.Errorf("Template not found for ID: " + filledTemplate.TemplateId.Hex())
		}

		var answers []*model.AnswersData
		for _, ans := range filledTemplate.Answers {
			answers = append(answers, &model.AnswersData{
				Key:   ans.Key,
				Value: ans.Value,
			})
		}

		savedTemplateRes := &model.SavedTemplatesData{
			ID:         filledTemplate.Id.Hex(),
			TemplateID: filledTemplate.TemplateId.Hex(),
			Title:      template.Title,
			Template:   template.Template,
			Topic:      template.Topic,
			Answers:    answers,
		}

		savedTemplates = append(savedTemplates, savedTemplateRes)
	}

	totalCount, err := filledTemplateColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count filled templates")
	}

	response := model.SavedTemplatePaginationResponse{
		Total:          int(totalCount),
		PerPage:        limit,
		CurrentPage:    page,
		TotalPages:     int(math.Ceil(float64(totalCount) / float64(limit))),
		SavedTemplates: savedTemplates,
	}

	return &response, nil
}
