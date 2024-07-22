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
	"go.mongodb.org/mongo-driver/mongo"
)

func EditCustomer(ctx context.Context, db *database.DB, input model.EditProfileRequestInput) (*model.Response, error) {
	var (
		customerColl = db.GetCollection("customer")
	)

	userData, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": userData.Id}

	var imageURL string
	if input.NewProfileImageFile != nil {
		file := input.NewProfileImageFile.File
		fileHeader := input.NewProfileImageFile.Filename

		fileExtension := filepath.Ext(fileHeader)
		id := primitive.NewObjectID()
		fileName := fmt.Sprintf("customer/%v-profilepic%v", id.Hex(), fileExtension)

		imageURL, err = utils.UploadToS3(fileName, file)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to upload image")
		}
	} else {
		imageURL = input.OldProfileImageURL
	}

	update := bson.M{
		"$set": bson.M{
			"name":      input.Name,
			"image":     imageURL,
			"updatedAt": time.Now().UTC(),
		},
	}

	session, err := db.GetMongoClient().StartSession()
	if err != nil {
		return nil, gqlerror.Errorf("Failed to start session")
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {

		updateRes, err := customerColl.UpdateOne(sessCtx, filter, update)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to update user")
		}

		if updateRes.MatchedCount == 0 {
			return nil, gqlerror.Errorf("No User found")
		}

		userFilter := bson.M{"user.id": userData.Id}
		ownGameUpdate := bson.M{
			"$set": bson.M{
				"user.image": imageURL,
				"user.name":  input.Name,
				"updatedAt":  time.Now().UTC(),
			},
		}

		_, err = db.GetCollection("ownGame").UpdateMany(sessCtx, userFilter, ownGameUpdate)
		if err != nil && err != mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Failed to update templates with the customer")
		}

		// friendFilter := bson.M{"friend.id": userData.Id}
		// friendUpdate := bson.M{
		// 	"$set": bson.M{
		// 		"image":     imageURL,
		// 		"updatedAt": time.Now().UTC(),
		// 	},
		// }

		// _, err = db.GetCollection("ownGame").UpdateMany(sessCtx, friendFilter, friendUpdate)
		// if err != nil && err != mongo.ErrNoDocuments {
		// 	return nil, gqlerror.Errorf("Failed to update templates with the friend")
		// }

		return nil, nil
	}

	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		return nil, gqlerror.Errorf("Transaction failed: %v", err)
	}

	return &model.Response{
		Message: "Profile Updated Successfully",
	}, nil
}

// curl --location 'http://localhost:5070/query' \
// --header 'X-GraphQL-Operation-Name: EditCustomer' \
// --header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJZCI6IjY2ODI1MmJkMDA0YjkxZGIyOTZiNGEwYiIsImVtYWlsIjoiY2hpcmFnc2hhcm1hNzU3NTdAZ21haWwuY29tIiwiZXhwIjoxNzM1NDcxMjAzLCJyb2xlIjoiY3VzdG9tZXIifQ.PbsCuq9TRp3NXumqtRIgB4_v6VWwup5jZrZCjDoNn_k' \
// --form 'operations="{\"query\":\"mutation EditProfile(\$input: EditProfileRequestInput\!) { editProfile(input: \$input) { message } }\",\"variables\":{\"input\":{\"name\":\"John Doe\",\"oldProfileImageUrl\":\"\"}}}"' \
// --form 'map="{\"profileImageFile\": [\"variables.input.newProfileImageFile\"]}"' \
// --form 'profileImageFile=@"/C:/Users/Chirag Sharma/Downloads/667bec9692362abd2c870718-image-Steve+Picture+3.jpg"'

// curl --location 'http://localhost:5070/query' \
// --header 'X-GraphQL-Operation-Name: EditCustomer' \
// --header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJZCI6IjY2ODI1MmJkMDA0YjkxZGIyOTZiNGEwYiIsImVtYWlsIjoiY2hpcmFnc2hhcm1hNzU3NTdAZ21haWwuY29tIiwiZXhwIjoxNzM1NDcxMjAzLCJyb2xlIjoiY3VzdG9tZXIifQ.PbsCuq9TRp3NXumqtRIgB4_v6VWwup5jZrZCjDoNn_k' \
// --form 'operations="{\"query\":\"mutation EditProfile(\$input: EditProfileRequestInput\!) { editProfile(input: \$input) { message } }\",\"variables\":{\"input\":{\"name\":\"John Doe\",\"oldProfileImageUrl\":\"https://laughifi-s3-bucket-dev.s3.us-east-1.amazonaws.com/customer/6683e52d0a32942d3b6c676c-profilepic.jpg\",\"newProfileImageFile\":null}}}"' \
// --form 'map="{}"'
