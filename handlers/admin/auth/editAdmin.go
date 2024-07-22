package auth

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/graph/model"
	"laughifi/utils"
	"path/filepath"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AdminEditAdmin(ctx context.Context, db *database.DB, input model.AdminEditProfileRequestInput) (*model.Response, error) {
	var (
		adminColl = db.GetCollection("admin")
	)

	adminData, err := utils.ExtractAdminFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": adminData.Id}

	var imageURL string
	if input.NewProfileImageFile != nil {
		file := input.NewProfileImageFile.File

		fileHeader := input.NewProfileImageFile.Filename

		fileExtension := filepath.Ext(fileHeader)

		id := primitive.NewObjectID()
		fileName := fmt.Sprintf("admin/%v-profilepic%v", id.Hex(), fileExtension)

		imageURL, err = utils.UploadToS3(fileName, file)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to upload image")
		}
	} else {
		imageURL = input.OldProfileImageURL
	}

	update := bson.M{
		"$set": bson.M{
			"firstName": input.FirstName,
			"lastName":  input.LastName,
			"image":     imageURL,
			"updatedAt": time.Now().UTC(),
		},
	}

	updateRes, err := adminColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update admin")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("No admin found")
	}

	return &model.Response{
		Message: "Profile Updated Successfully",
	}, nil
}
