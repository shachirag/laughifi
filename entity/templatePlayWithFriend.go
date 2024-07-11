package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TemplatePlayWithFriendEntity struct {
	Id         primitive.ObjectID                    `json:"id" bson:"_id"`
	FriendId   primitive.ObjectID                    `json:"friendId" bson:"friendId"`
	UserId     primitive.ObjectID                    `json:"userId" bson:"userId"`
	Status     string                                `json:"status" bson:"status"`
	TemplateId primitive.ObjectID                    `json:"templateId" bson:"templateId"`
	Answers    []AnswersTemplatePlayWithFriendEntity `json:"answers" bson:"answers"`
	CreatedAt  time.Time                             `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time                             `json:"updatedAt" bson:"updatedAt"`
}

type AnswersTemplatePlayWithFriendEntity struct {
	Id    primitive.ObjectID `json:"id" bson:"id"`
	Key   string             `json:"key" bson:"key"`
	Value string             `json:"value" bson:"value"`
}
