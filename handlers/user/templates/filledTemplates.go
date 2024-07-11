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
		return nil, gqlerror.Errorf("invalid template Id")
	}

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	if data.Answers == nil || len(data.Answers) == 0 {
		return nil, gqlerror.Errorf("At least one answer is required to share")
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

// {
//     "query": "mutation FilledTemplate($input: FilledTemplateRequestInput!) { filledTemplate(input: $input) { message } }",
//     "variables": {
//         "input": {
//             "templateId": "66823dc7b483e6dedf106c7d",
//             "answers": [
//                 {
//                     "key": "BLANK1",
//                     "value": "ANSWER_VALUE_1"
//                 },
//                 {
//                     "key": "BLANK2",
//                     "value": "ANSWER_VALUE_2"
//                 },
//                 {
//                     "key": "BLANK3",
//                     "value": "ANSWER_VALUE_3"
//                 },
//                 {
//                     "key": "BLANK4",
//                     "value": "ANSWER_VALUE_4"
//                 },
//                 {
//                     "key": "BLANK5",
//                     "value": "ANSWER_VALUE_5"
//                 },
//                 {
//                     "key": "BLANK6",
//                     "value": "ANSWER_VALUE_6"
//                 },
//                 {
//                     "key": "BLANK7",
//                     "value": "ANSWER_VALUE_7"
//                 }
//             ]
//         }
//     }
// }
