package trivia

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func EditTrivia(ctx context.Context, db *database.DB, triviaId string, input model.EditTriviaRequestInput) (*model.Response, error) {
	var (
		triviaColl = db.GetCollection("trivia")
	)

	triviaObjID, err := primitive.ObjectIDFromHex(triviaId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid trivia Id")
	}

	categoryObjID, err := primitive.ObjectIDFromHex(input.CategoryID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid category Id")
	}

	var category entity.CategoryEntity
	err = db.GetCollection("category").FindOne(ctx, bson.M{"_id": categoryObjID}).Decode(&category)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch category")
	}

	var dareForYouUrl string
	if input.NewDareForWrongAnswerFile != nil {
		file := input.NewDareForWrongAnswerFile.File
		id := primitive.NewObjectID()
		fileName := fmt.Sprintf("dare/%v.jpg", id.Hex())

		dareForYouUrl, err = utils.UploadToS3(fileName, file)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to upload dareForYou")
		}
	} else {
		dareForYouUrl = input.OldDareForWrongAnswerURL
	}

	filter := bson.M{"_id": triviaObjID}

	update := bson.M{
		"$set": bson.M{
			"question":           input.Question,
			"answers":            input.Answers,
			"category.id":        categoryObjID,
			"category.name":      category.Name,
			"dareForWrongAnswer": dareForYouUrl,
			"correctAnswer":      input.Answers,
			"updatedAt":          time.Now().UTC(),
		},
	}

	updateRes, err := triviaColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update trivia")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("No trivia found")
	}

	return &model.Response{
		Message: "Trivia Updated Successfully",
	}, nil
}
