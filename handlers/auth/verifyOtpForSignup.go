package auth

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"os"
	"strings"
	"time"

	jtoken "github.com/golang-jwt/jwt/v4"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func VerifyOtpForSignup(ctx context.Context, db *database.DB, data model.VerifyOtpForSignUpRequestInput) (*model.LoginResponse, error) {
	var (
		otpColl      = db.GetCollection("otp")
		customerColl = db.GetCollection("customer")
		otpData      entity.OtpEntity
		customer     entity.CustomerEntity
	)

	if data.Otp == "" {
		return nil, gqlerror.Errorf("Otp is required")
	}

	smallEmail := strings.ToLower(data.Email)

	otpFilter := bson.M{"email": smallEmail}

	err := otpColl.FindOne(ctx, otpFilter, options.FindOne().SetSort(bson.M{"createdAt": -1})).Decode(&otpData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Invalid Otp")
		}

		return nil, gqlerror.Errorf("Internal Server Error")
	}

	if data.Otp != otpData.Otp {
		return nil, gqlerror.Errorf("Invalid Otp")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to hash password")
	}

	filter := bson.M{
		"email": strings.ToLower(data.Email),
	}

	exists, err := customerColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to count customers")
	}

	if exists > 0 {
		return nil, gqlerror.Errorf("Email is already in use")
	}

	id := primitive.NewObjectID()

	customer = entity.CustomerEntity{
		Id:        id,
		Name:      data.Name,
		Email:     smallEmail,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err = customerColl.InsertOne(ctx, customer)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to insert customer")
	}

	var hasPassword bool
	if customer.Password == "" {
		hasPassword = false
	} else {
		hasPassword = true
	}

	_secret := os.Getenv("JWT_SECRET_KEY")

	month := (time.Hour * 24) * 30
	claims := jtoken.MapClaims{
		"Id":    customer.Id,
		"email": smallEmail,
		"role":  "customer",
		"exp":   time.Now().Add(month * 6).Unix(),
	}

	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	_token, err := token.SignedString([]byte(_secret))
	if err != nil {
		return nil, gqlerror.Errorf("Token is not valid")
	}

	return &model.LoginResponse{
		ID:          customer.Id.Hex(),
		Name:        customer.Name,
		Email:       customer.Email,
		Image:       customer.Image,
		CreatedAt:   customer.CreatedAt.Format(time.DateTime),
		UpdatedAt:   customer.UpdatedAt.Format(time.DateTime),
		Token:       _token,
		HasPassword: hasPassword,
	}, nil
}
