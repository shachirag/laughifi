package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OwnGameEntity struct {
	Id        primitive.ObjectID  `json:"id" bson:"_id"`
	Friends   []Friend            `json:"friends" bson:"friends"`
	User      User                `json:"user" bson:"user"`
	Status    string              `json:"status" bson:"status"`
	GameName  string              `json:"gameName" bson:"gameName"`
	Category  Category            `json:"category" bson:"category"`
	TriviaID  *primitive.ObjectID `json:"triviaId,omitempty" bson:"triviaId,omitempty"`
	CreatedAt time.Time           `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt" bson:"updatedAt"`
}

type User struct {
	Id    primitive.ObjectID `json:"id" bson:"id"`
	Name  string             `json:"name" bson:"name"`
	Image string             `json:"image" bson:"image"`
}

type Friend struct {
	Id              primitive.ObjectID `json:"id" bson:"id"`
	Image           string             `json:"image" bson:"image"`
	Name            string             `json:"name" bson:"name"`
	IsCorrectAnswer *bool              `json:"isCorrectAnswer,omitempty" bson:"isCorrectAnswer,omitempty"`
	Answer          *string            `json:"answer,omitempty" bson:"answer,omitempty"`
	PlayedStatus    string             `json:"playedStatus" bson:"playedStatus"`
}
