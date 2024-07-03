package utils

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ExtractClaimsFromContext(ctx context.Context) (jwt.MapClaims, error) {
	claims, ok := ctx.Value(middleware.UserClaimsKey).(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("no claims found in context")
	}
	return claims, nil
}

func ExtractUserFromContext(ctx context.Context, db *database.DB) (*entity.CustomerEntity, error) {
	claims, err := ExtractClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userIDHex, ok := claims["Id"].(string)
	if !ok {
		return nil, fmt.Errorf("user ID not found in token claims")
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format in token claims")
	}

	var user entity.CustomerEntity
	customerColl := db.GetCollection("customer")
	err = customerColl.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to fetch user details: "+err.Error())
	}

	return &user, nil
}

func ExtractAdminFromContext(ctx context.Context, db *database.DB) (*entity.AdminEntity, error) {
	claims, err := ExtractClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userIDHex, ok := claims["Id"].(string)
	if !ok {
		return nil, fmt.Errorf("admin ID not found in token claims")
	}

	adminID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return nil, fmt.Errorf("invalid admin ID format in token claims")
	}

	var admin entity.AdminEntity
	adminColl := db.GetCollection("admin")
	err = adminColl.FindOne(ctx, bson.M{"_id": adminID}).Decode(&admin)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to fetch admin details: "+err.Error())
	}

	return &admin, nil
}