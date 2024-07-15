package category

import (
	"context"
	"laughifi/database"
	"laughifi/graph/model"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func EditCategory(ctx context.Context, db *database.DB, categoryId string, input model.EditCategoryRequestInput) (*model.Response, error) {
	var (
		categoryColl = db.GetCollection("category")
	)

	categoryObjID, err := primitive.ObjectIDFromHex(categoryId)
	if err != nil {
		return nil, gqlerror.Errorf("invalid category Id")
	}

	filter := bson.M{"_id": categoryObjID}

	session, err := db.GetMongoClient().StartSession()
	if err != nil {
		return nil, gqlerror.Errorf("Failed to start session")
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {

		update := bson.M{
			"$set": bson.M{
				"name":      input.Category,
				"updatedAt": time.Now().UTC(),
			},
		}

		updateRes, err := categoryColl.UpdateOne(sessCtx, filter, update)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to update category")
		}

		if updateRes.MatchedCount == 0 {
			return nil, gqlerror.Errorf("No category found")
		}

		templatesFilter := bson.M{"category.id": categoryObjID}
		templatesUpdate := bson.M{
			"$set": bson.M{
				"category.name": input.Category,
				"updatedAt":     time.Now().UTC(),
			},
		}

		_, err = db.GetCollection("template").UpdateMany(sessCtx, templatesFilter, templatesUpdate)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to update templates with the new category name")
		}

		ownGameFilter := bson.M{"category.id": categoryObjID}
		ownGameUpdate := bson.M{
			"$set": bson.M{
				"category.name": input.Category,
				"updatedAt":     time.Now().UTC(),
			},
		}

		_, err = db.GetCollection("ownGame").UpdateMany(sessCtx, ownGameFilter, ownGameUpdate)
		if err != nil && err != mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Failed to update templates with the new category name")
		}

		return nil, nil
	}

	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		return nil, gqlerror.Errorf("Transaction failed: %v", err)
	}

	return &model.Response{
		Message: "Category Updated Successfully",
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
