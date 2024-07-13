package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionEntity struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	UserId    primitive.ObjectID `json:"userId" bson:"userId"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}
