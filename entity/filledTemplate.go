package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FilledTemplateEntity struct {
	Id         primitive.ObjectID `json:"id" bson:"_id"`
	UserId     primitive.ObjectID `json:"userId" bson:"userId"`
	TemplateId primitive.ObjectID `json:"templateId" bson:"templateId"`
	Answers    []Answers          `json:"answers" bson:"answers"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type Answers struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}
