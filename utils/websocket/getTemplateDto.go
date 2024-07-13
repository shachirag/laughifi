package websocket

import (
	"laughifi/database"
	"laughifi/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetTemplateDto(db *database.DB, templateId primitive.ObjectID) *TemplateServerToClientDto {

	var playWithFriendTemplate entity.TemplatePlayWithFriendEntity

	filter := bson.M{
		"_id": templateId,
	}

	err := db.GetCollection("templatePlayWithFriend").FindOne(ctx, filter).Decode(&playWithFriendTemplate)
	if err != nil {
		return nil
	}

	var template entity.TemplatesEntity

	templateFilter := bson.M{
		"_id": playWithFriendTemplate.TemplateId,
	}

	err = db.GetCollection("template").FindOne(ctx, templateFilter).Decode(&template)
	if err != nil {
		return nil
	}

	var answers []Answers
	for _, ans := range playWithFriendTemplate.Answers {
		answers = append(answers, Answers{
			Key:   ans.Key,
			Value: ans.Value,
		})
	}

	newTemplateObj := TemplateServerToClientDto{
		Id:       playWithFriendTemplate.Id,
		Template: template.Template,
		Category: template.Category.Name,
		Title:    template.Title,
		Status:   playWithFriendTemplate.Status,
		Answers:  answers,
	}

	return &newTemplateObj
}

type TemplateServerToClientDto struct {
	Id       primitive.ObjectID `json:"sendAnswerUserId"`
	Template string             `json:"template"`
	Title    string             `json:"title"`
	Category string             `json:"category"`
	Status   string             `json:"status"`
	Answers  []Answers          `json:"answers"`
}

type Answers struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
