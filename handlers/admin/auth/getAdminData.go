package auth

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAdminData(ctx context.Context, db *database.DB) (*model.AdminLoginResponse, error) {
	var admin entity.AdminEntity

	adminData, err := utils.ExtractAdminFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	adminColl := db.GetCollection("admin")

	err = adminColl.FindOne(ctx, bson.M{"_id": adminData.Id}).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("admin not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch admin")
	}

	return &model.AdminLoginResponse{
		ID:        admin.Id.Hex(),
		FirstName: admin.FirstName,
		LastName:  admin.LastName,
		Email:     admin.Email,
		Image:     admin.Image,
		CreatedAt: admin.CreatedAt.Format(time.DateTime),
		UpdatedAt: admin.UpdatedAt.Format(time.DateTime),
		Token:     "",
	}, nil
}
