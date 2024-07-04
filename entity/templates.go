package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TemplatesEntity struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	Category  Category           `json:"category" bson:"category"`
	Title     string             `json:"title" bson:"title"`
	Template  string             `json:"template" bson:"template"`
	IsDeleted bool               `json:"isDeleted" bson:"isDeleted"`
	CreatedAt time.Time          `json:"cresatedAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type Category struct {
	Id   primitive.ObjectID `json:"id" bson:"id"`
	Name string             `json:"name" bson:"name"`
}
