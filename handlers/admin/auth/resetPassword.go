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

func AdminResetPassword(ctx context.Context, db *database.DB, input model.ResetPasswordRequestInput) (*model.Response, error) {
	var (
		adminColl = db.GetCollection("admin")
		admin     entity.AdminEntity
	)

	smallEmail := strings.ToLower(input.Email)

	err := adminColl.FindOne(ctx, bson.M{"email": smallEmail}).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("No admin found with the provided email.")
		}
		return nil, gqlerror.Errorf("Internal server error while fetching the admin: " + err.Error())
	}

	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to hash the password: " + err.Error())
	}

	_, err = adminColl.UpdateOne(ctx, bson.M{"_id": admin.Id}, bson.M{"$set": bson.M{"password": hashedPassword}})
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update the password in the database: " + err.Error())
	}

	return &model.Response{
		Message: "Password Reset Successfully",
	}, nil
}
