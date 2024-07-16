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

func GetFriendTemplates(ctx context.Context, db *database.DB, page int, limit int, friendId string) (*model.FriendTemplatePaginationResponse, error) {

	friendObjID, err := primitive.ObjectIDFromHex(friendId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid friend Id")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}

	templatePlayWithFriendColl := db.GetCollection("templatePlayWithFriend")

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"shareIds": bson.M{
			"$all": []primitive.ObjectID{user.Id, friendObjID},
		},
		// "userId":   user.Id,
		// "friendId": friendObjID,
		"isFriend": true,
	}

	skip := (page - 1) * limit
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"updatedAt": -1})

	cursor, err := templatePlayWithFriendColl.Find(ctx, filter, findOptions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.FriendTemplatePaginationResponse{
				Total:           0,
				PerPage:         limit,
				CurrentPage:     page,
				TotalPages:      0,
				FriendTemplates: []*model.FriendTemplates{},
			}, nil
		}
	}
	defer cursor.Close(ctx)

	var friendTemplatess []entity.TemplatePlayWithFriendEntity
	var templateIds []primitive.ObjectID

	for cursor.Next(ctx) {
		var friendTemplate entity.TemplatePlayWithFriendEntity
		err := cursor.Decode(&friendTemplate)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to decode friend template")
		}
		friendTemplatess = append(friendTemplatess, friendTemplate)
		templateIds = append(templateIds, friendTemplate.TemplateId)
	}

	if err := cursor.Err(); err != nil {
		return nil, gqlerror.Errorf("Cursor error: " + err.Error())
	}

	if len(templateIds) > 0 {
		templateFilter := bson.M{"_id": bson.M{"$in": templateIds}}

		templateColl := db.GetCollection("template")
		templateCursor, err := templateColl.Find(ctx, templateFilter)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to fetch templates")
		}
		defer templateCursor.Close(ctx)

		templateMap := make(map[primitive.ObjectID]entity.TemplatesEntity)
		for templateCursor.Next(ctx) {
			var template entity.TemplatesEntity
			err := templateCursor.Decode(&template)
			if err != nil {
				return nil, gqlerror.Errorf("Failed to decode template")
			}
			templateMap[template.Id] = template
		}

		if err := templateCursor.Err(); err != nil {
			return nil, gqlerror.Errorf("Cursor error while fetching templates: " + err.Error())
		}

		var friendTemplates []*model.FriendTemplates
		for _, friendTemplate := range friendTemplatess {
			template, found := templateMap[friendTemplate.TemplateId]
			if !found {
				continue
			}

			var answers []*model.FriendAnswers
			for _, ans := range friendTemplate.Answers {
				answers = append(answers, &model.FriendAnswers{
					Key:   ans.Key,
					Value: ans.Value,
				})
			}

			friendTemplateRes := &model.FriendTemplates{
				ID:       friendTemplate.Id.Hex(),
				Title:    template.Title,
				Template: template.Template,
				Category: template.Category.Name,
				Answers:  answers,
				Status:   friendTemplate.Status,
			}

			friendTemplates = append(friendTemplates, friendTemplateRes)
		}

		totalCount, err := templatePlayWithFriendColl.CountDocuments(ctx, filter)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to count friend templates")
		}

		response := model.FriendTemplatePaginationResponse{
			Total:           int(totalCount),
			PerPage:         limit,
			CurrentPage:     page,
			TotalPages:      int(math.Ceil(float64(totalCount) / float64(limit))),
			FriendTemplates: friendTemplates,
		}

		return &response, nil
	}

	return &model.FriendTemplatePaginationResponse{
		Total:           0,
		PerPage:         limit,
		CurrentPage:     page,
		TotalPages:      0,
		FriendTemplates: []*model.FriendTemplates{},
	}, nil
}
