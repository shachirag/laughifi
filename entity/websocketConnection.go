package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebsocketConnectionEntity struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	ConnectionId string             `json:"connectionId" bson:"connectionId"`
	RoleId       primitive.ObjectID `json:"roleId" bson:"roleId"`
	Role         string             `json:"role" bson:"role"`
}
