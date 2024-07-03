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

func GetUserData(ctx context.Context, db *database.DB) (*model.LoginResponse, error) {
	var customer entity.CustomerEntity

	userData, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	customerColl := db.GetCollection("customer")

	err = customerColl.FindOne(ctx, bson.M{"_id": userData.Id}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("User not Found")
		}
		return nil, gqlerror.Errorf("Failed to fetch user")
	}

	var hasPassword bool
	if customer.Password == "" {
		hasPassword = false
	} else {
		hasPassword = true
	}

	return &model.LoginResponse{
		ID:          customer.Id.Hex(),
		Name:        customer.Name,
		Email:       customer.Email,
		Image:       customer.Image,
		CreatedAt:   customer.CreatedAt.Format(time.DateTime),
		UpdatedAt:   customer.UpdatedAt.Format(time.DateTime),
		Token:       "",
		HasPassword: hasPassword,
	}, nil
}
