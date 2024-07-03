package auth

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/graph/model"
	"laughifi/utils"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func EditAdmin(ctx context.Context, db *database.DB, input model.AdminEditProfileRequestInput) (*model.Response, error) {
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
		id := primitive.NewObjectID()
		fileName := fmt.Sprintf("admin/%v-profilepic.jpg", id.Hex())

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
