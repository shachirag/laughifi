package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminEntity struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	FirstName string             `json:"firstName" bson:"firstName"`
	LastName  string             `json:"lastName" bson:"lastName"`
	Email     string             `json:"email" bson:"email"`
	Password  string             `json:"password" bson:"password"`
	Image     string             `json:"image" bson:"image"`
	CreatedAt time.Time          `json:"cresatedAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}
