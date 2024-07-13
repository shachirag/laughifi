package middleware

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MiddlewareWebsocketData struct {
	ID   primitive.ObjectID
	Role string
}

func SetWebSocketMiddlewareData(tokenString string) (*MiddlewareWebsocketData, error) {

	jwtSecret:=os.Getenv("JWT_SECRET_KEY")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("token claims are not of type jwt.MapClaims")
	}

	cId, ok := claims["Id"].(string)
	if !ok {
		return nil, errors.New("ID claim not found or not a string")
	}
	id, err := primitive.ObjectIDFromHex(cId)
	if err != nil {
		return nil, err
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("role claim not found or not a string")
	}

	middlewareData := &MiddlewareWebsocketData{
		ID:   id,
		Role: role,
	}

	return middlewareData, nil
}
