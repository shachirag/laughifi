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
	"golang.org/x/crypto/bcrypt"
)

func Signup(ctx context.Context, db *database.DB, data model.SignUpRequestInput) (*model.LoginResponse, error) {
	var (
		customerColl = db.GetCollection("customer")
		customer     entity.CustomerEntity
	)

	smallEmail := strings.ToLower(data.Email)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to hash password")
	}

	filter := bson.M{
		"email":     smallEmail,
		"isDeleted": false,
	}

	exists, err := customerColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("failed to count customers")
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
		IsDeleted: false,
	}

	_, err = customerColl.InsertOne(ctx, customer)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to insert user")
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
		ID:        id.Hex(),
		Name:      data.Name,
		Email:     data.Email,
		Image:     customer.Image,
		CreatedAt: time.Now().UTC().Format(time.DateTime),
		UpdatedAt: time.Now().UTC().Format(time.DateTime),
		Token:     _token,
		HasPassword:hasPassword,
	}, nil
}
