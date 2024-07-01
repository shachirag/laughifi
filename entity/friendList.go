package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FriendListEntity struct {
	Id          primitive.ObjectID `json:"id" bson:"_id"`
	UserId      primitive.ObjectID `json:"userId" bson:"userId"`
	FriendsList []FriendsList      `json:"friendsList" bson:"friendsList"`
	CreatedAt   time.Time          `json:"cresatedAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type FriendsList struct {
	Id     primitive.ObjectID `json:"id" bson:"id"`
	Status string             `json:"status"`
}
