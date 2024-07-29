package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebsocketConnectionEntity struct {
	Id           primitive.ObjectID  `json:"id" bson:"_id"`
	ConnectionId string              `json:"connectionId" bson:"connectionId"`
	RoleId       primitive.ObjectID  `json:"roleId" bson:"roleId"`
	GameType     *string             `json:"gameType,omitempty" bson:"gameType,omitempty"`
	TriviaId     *primitive.ObjectID `json:"triviaId,omitempty" bson:"triviaId,omitempty"`
	TemplateId   *primitive.ObjectID `json:"templateId,omitempty" bson:"templateId,omitempty"`
	Role         string              `json:"role" bson:"role"`
}
