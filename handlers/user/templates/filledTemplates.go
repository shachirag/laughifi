package templates

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func FilledTemplates(ctx context.Context, db *database.DB, data model.FilledTemplateRequestInput) (*model.Response, error) {
	var (
		filledTemplateColl = db.GetCollection("filledTemplate")
		filledTemplate     entity.FilledTemplateEntity
	)

	templateObjId, err := primitive.ObjectIDFromHex(data.TemplateID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid user Id")
	}

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	var answers []entity.Answers
	if data.Answers != nil {
		for _, answerInput := range data.Answers {
			answer := entity.Answers{
				Key:   answerInput.Key,
				Value: answerInput.Value,
			}
			answers = append(answers, answer)
		}
	}

	id := primitive.NewObjectID()

	filledTemplate = entity.FilledTemplateEntity{
		Id:         id,
		UserId:     user.Id,
		TemplateId: templateObjId,
		Answers:    answers,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	_, err = filledTemplateColl.InsertOne(ctx, filledTemplate)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to save template")
	}

	return &model.Response{
		Message: "Template Saved Successfully",
	}, nil
}
