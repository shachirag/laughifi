package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WouldYouRatherEntity struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	Question  string             `json:"question" bson:"question"`
	Options   []string           `json:"options" bson:"options"`
	IsDeleted bool               `json:"isDeleted" bson:"isDeleted"`
	CreatedAt time.Time          `json:"cresatedAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}
