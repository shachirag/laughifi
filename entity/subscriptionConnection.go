package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionEntity struct {
	Id         primitive.ObjectID `json:"id" bson:"_id"`
	UserId     primitive.ObjectID `json:"userId" bson:"userId"`
	TemplateId primitive.ObjectID `json:"templateId" bson:"templateId"`
	Answers    []Answers          `json:"answers" bson:"answers"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
}
