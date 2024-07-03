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
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func LoginCustomer(ctx context.Context, db *database.DB, input model.LoginRequestInput) (*model.LoginResponse, error) {
	customerColl := db.GetCollection("customer")

	smallEmail := strings.ToLower(input.Email)

	filter := bson.M{"email": smallEmail}

	var customer entity.CustomerEntity
	err := customerColl.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Invalid credentials")
		}
		return nil, gqlerror.Errorf("Error occurred while fetching user: " + err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(strings.TrimSpace(input.Password)))
	if err != nil {
		return nil, gqlerror.Errorf("Invalid credentials")
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
