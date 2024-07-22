package auth

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"os"
	"strings"
	"time"

	jtoken "github.com/golang-jwt/jwt/v4"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SocialLoginCustomer(ctx context.Context, db *database.DB, input model.SocialLoginRequestInput) (*model.LoginResponse, error) {
	var (
		customerColl = db.GetCollection("customer")
		customer     *entity.CustomerEntity
	)

	err := utils.ValidateSocialId(&input)
	if err != nil {
		return nil, gqlerror.Errorf("Invalid social ID: " + err.Error())
	}

	filter := bson.M{}
	switch input.Type {
	case "apple":
		filter = bson.M{"socialDetails.appleId": input.SocialID, "isDeleted": false}
	case "google":
		filter = bson.M{"socialDetails.googleId": input.SocialID, "isDeleted": false}
	default:
		return nil, gqlerror.Errorf("Unsupported social login type")
	}

	var smallEmail string
	if input.Email != nil {
		smallEmail = strings.ToLower(*input.Email)
	}

	err = customerColl.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments && input.Email != nil {
			filter = bson.M{"email": smallEmail, "isDeleted": false}
			err = customerColl.FindOne(ctx, filter).Decode(&customer)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					customer, err = socialSignup(ctx, db, &input)
					if err != nil {
						return nil, gqlerror.Errorf("Failed to sign up: " + err.Error())
					}
				} else {
					return nil, gqlerror.Errorf("Error finding user by email: " + err.Error())
				}
			}
		} else {
			return nil, gqlerror.Errorf("User not found with provided social ID and no email to fallback")
		}
	}

	update := bson.M{
		"deviceInfo.deviceToken": input.DeviceInfo.DeviceToken,
		"deviceInfo.deviceType":  input.DeviceInfo.DeviceType,
	}

	if customer != nil {

		isUpdate := false
		if customer.Email == "" && smallEmail != "" {
			update["email"] = smallEmail
			isUpdate = true
		}
		if customer.SocialDetails.AppleId == "" && input.Type == "apple" {
			update["socialDetails.appleId"] = input.SocialID
			isUpdate = true
		}
		if customer.SocialDetails.GoogleId == "" && input.Type == "google" {
			update["socialDetails.googleId"] = input.SocialID
			isUpdate = true
		}

		if isUpdate {
			_, err = customerColl.UpdateOne(ctx, bson.M{"_id": customer.Id}, bson.M{"$set": update})
			if err != nil {
				return nil, gqlerror.Errorf("Failed to update user details: " + err.Error())
			}
		}
	}

	deviceInfoUpdate := bson.M{
		"deviceInfo.deviceToken": input.DeviceInfo.DeviceToken,
		"deviceInfo.deviceType":  input.DeviceInfo.DeviceType,
	}

	_, err = customerColl.UpdateOne(ctx, bson.M{"_id": customer.Id}, bson.M{"$set": deviceInfoUpdate})
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update user details: " + err.Error())
	}

	var hasPassword bool
	if customer.Password == "" {
		hasPassword = false
	} else {
		hasPassword = true
	}

	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return nil, gqlerror.Errorf("JWT secret key not found.")
	}

	claims := jtoken.MapClaims{
		"Id":    customer.Id,
		"email": customer.Email,
		"role":  "customer",
		"exp":   time.Now().Add(6 * 30 * 24 * time.Hour).Unix(),
	}

	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, gqlerror.Errorf("Failed to generate JWT token: " + err.Error())
	}

	return &model.LoginResponse{
		ID:          customer.Id.Hex(),
		Name:        customer.Name,
		Email:       customer.Email,
		Image:       customer.Image,
		CreatedAt:   customer.CreatedAt.Format(time.DateTime),
		UpdatedAt:   customer.UpdatedAt.Format(time.DateTime),
		Token:       signedToken,
		HasPassword: hasPassword,
	}, nil
}

func socialSignup(ctx context.Context, db *database.DB, data *model.SocialLoginRequestInput) (*entity.CustomerEntity, error) {
	userColl := db.GetCollection("customer")

	var smallEmail string
	if data.Email != nil {
		smallEmail = strings.ToLower(*data.Email)
	}

	filter := bson.M{"email": smallEmail, "isDeleted": false}
	exists, err := userColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Error checking existing user: " + err.Error())
	}
	if exists > 0 {
		return nil, gqlerror.Errorf("User with this email already exists")
	}

	id := primitive.NewObjectID()

	customer := &entity.CustomerEntity{
		Id:    id,
		Email: smallEmail,
		Name:  data.Name,
		DeviceInfo: entity.DeviceInfo{
			DeviceToken: data.DeviceInfo.DeviceToken,
			DeviceType:  data.DeviceInfo.DeviceType,
		},
		IsDeleted: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	switch data.Type {
	case "google":
		customer.SocialDetails.GoogleId = data.SocialID
	case "apple":
		customer.SocialDetails.AppleId = data.SocialID
	default:
		return nil, gqlerror.Errorf("Unsupported social login type")
	}

	_, err = userColl.InsertOne(ctx, customer)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to insert new customer: " + err.Error())
	}

	return customer, nil
}
