package auth

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"strings"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ResetPassword(ctx context.Context, db *database.DB, input model.ResetPasswordRequestInput) (*model.Response, error) {
	var (
		customerColl = db.GetCollection("customer")
		user         entity.CustomerEntity
	)

	smallEmail := strings.ToLower(input.Email)

	err := customerColl.FindOne(ctx, bson.M{"email": smallEmail}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("No user found with the provided email.")
		}
		return nil, gqlerror.Errorf("Internal server error while fetching the user: " + err.Error())
	}

	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to hash the password: " + err.Error())
	}

	_, err = customerColl.UpdateOne(ctx, bson.M{"_id": user.Id}, bson.M{"$set": bson.M{"password": hashedPassword}})
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update the password in the database: " + err.Error())
	}

	return &model.Response{
		Message: "Password Reset Successfully",
	}, nil
}
