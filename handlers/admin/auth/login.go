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

func LoginAdmin(ctx context.Context, db *database.DB, input model.AdminLoginRequestInput) (*model.AdminLoginResponse, error) {
	adminColl := db.GetCollection("admin")

	smallEmail := strings.ToLower(input.Email)

	filter := bson.M{"email": smallEmail}

	var admin entity.AdminEntity
	err := adminColl.FindOne(ctx, filter).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Invalid credentials")
		}
		return nil, gqlerror.Errorf("Error occurred while fetching admin: " + err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(strings.TrimSpace(input.Password)))
	if err != nil {
		return nil, gqlerror.Errorf("Invalid credentials")
	}

	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return nil, gqlerror.Errorf("JWT secret key not found.")
	}

	claims := jtoken.MapClaims{
		"Id":    admin.Id,
		"email": admin.Email,
		"role":  "admin",
		"exp":   time.Now().Add(6 * 30 * 24 * time.Hour).Unix(),
	}

	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, gqlerror.Errorf("Failed to generate JWT token: " + err.Error())
	}

	return &model.AdminLoginResponse{
		ID:        admin.Id.Hex(),
		FirstName: admin.FirstName,
		LastName:  admin.LastName,
		Email:     admin.Email,
		Image:     admin.Image,
		CreatedAt: admin.CreatedAt.Format(time.DateTime),
		UpdatedAt: admin.UpdatedAt.Format(time.DateTime),
		Token:     signedToken,
	}, nil
}
