package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TemplateEntity struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	Title    string             `json:"title" bson:"title"`
	Topic    string             `json:"topic" bson:"topic"`
	Template string             `json:"template" bson:"template"`
}
