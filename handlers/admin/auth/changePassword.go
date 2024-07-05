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

func AdminChangePassword(ctx context.Context, db *database.DB, input model.ChangePasswordRequestInput) (*model.Response, error) {
	var (
		adminColl = db.GetCollection("admin")
	)

	adminData, err := utils.ExtractAdminFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": adminData.Id}

	result := adminColl.FindOne(ctx, filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Admin not found")
		}

		return nil, gqlerror.Errorf("Error finding admin")
	}

	var admin entity.AdminEntity
	err = result.Decode(&admin)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to decode admin")
	}

	if input.CurrentPassword != "" {
		err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(input.CurrentPassword))
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

	updateRes, err := adminColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update password")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("admin not found")
	}

	return &model.Response{
		Message: "Password Changed Successfully",
	}, nil
}
