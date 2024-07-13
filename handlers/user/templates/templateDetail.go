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

func GetTemplateDetail(ctx context.Context, db *database.DB, templateId string) (*model.FriendTemplateDetail, error) {
	var playWithFriendTemplate entity.TemplatePlayWithFriendEntity

	templateObjID, err := primitive.ObjectIDFromHex(templateId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid user Id")
	}

	templatePlayWithFriendColl := db.GetCollection("templatePlayWithFriend")

	err = templatePlayWithFriendColl.FindOne(ctx, bson.M{"_id": templateObjID}).Decode(&playWithFriendTemplate)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("template not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch template")
	}

	var template entity.TemplatesEntity

	templateFilter := bson.M{
		"_id": playWithFriendTemplate.TemplateId,
	}

	err = db.GetCollection("template").FindOne(ctx, templateFilter).Decode(&template)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch template")
	}

	var answers []*model.FriendAnswers
	for _, ans := range playWithFriendTemplate.Answers {
		answers = append(answers, &model.FriendAnswers{
			Key:   ans.Key,
			Value: ans.Value,
		})
	}

	return &model.FriendTemplateDetail{
		ID:       playWithFriendTemplate.Id.Hex(),
		Template: template.Template,
		Category: template.Category.Name,
		Title:    template.Title,
		Answers:  answers,
	}, nil
}
