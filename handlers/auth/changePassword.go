package auth

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func ChangePassword(ctx context.Context, db *database.DB, input model.ChangePasswordRequestInput) (*model.Response, error) {
	var (
		customerColl = db.GetCollection("customer")
	)

	userData, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": userData.Id}

	result := customerColl.FindOne(ctx, filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("User not found")
		}

		return nil, gqlerror.Errorf("Error finding user")
	}

	var user entity.CustomerEntity
	err = result.Decode(&user)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to decode user")
	}

	if input.CurrentPassword != "" {
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword))
		if err != nil {
			return nil, gqlerror.Errorf("Current Password is Incorrect")
		}
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 6)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to hash new password")
	}

	update := bson.M{
		"$set": bson.M{
			"password": string(hashedNewPassword),
		},
	}

	updateRes, err := customerColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update password")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("User not found")
	}
	return &model.Response{
		Message: "Password Changed Successfully",
	}, nil
}
