package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CategoryEntity struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	IsDeleted bool               `json:"isDeleted" bson:"isDeleted"`
	CreatedAt time.Time          `json:"cresatedAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}
