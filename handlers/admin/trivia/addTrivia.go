package trivia

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"path/filepath"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddTrivia(ctx context.Context, db *database.DB, data model.TriviaRequestInput) (*model.Response, error) {
	var (
		triviaColl = db.GetCollection("trivia")
		trivia     entity.TriviaEntity
		category   entity.CategoryEntity
	)

	categoryObjID, err := primitive.ObjectIDFromHex(data.CategoryID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid category Id")
	}

	err = db.GetCollection("category").FindOne(ctx, bson.M{"_id": categoryObjID}).Decode(&category)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch category")
	}

	id := primitive.NewObjectID()

	var dareForYouURL string
	if data.DareForWrongAnswerFile != nil {
		dareForYouFile := data.DareForWrongAnswerFile.File
		fileHeader := data.DareForWrongAnswerFile.Filename

		fileExtension := filepath.Ext(fileHeader)
		dareForYouFileName := fmt.Sprintf("dare/%v%v", id.Hex(), fileExtension)

		dareForYouURL, err = utils.UploadToS3(dareForYouFileName, dareForYouFile)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to upload dareForYou")
		}
	}

	trivia = entity.TriviaEntity{
		Id:                 id,
		Question:           data.Question,
		Answers:            data.Answers,
		CorrectAnswer:      data.CorrectAnswer,
		DareForWrongAnswer: dareForYouURL,
		IsDeleted:          false,
		Category: entity.Category{
			Id:   categoryObjID,
			Name: category.Name,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err = triviaColl.InsertOne(ctx, trivia)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to insert trivia")
	}

	return &model.Response{
		Message: "Trivia Saved Successfully",
	}, nil
}
