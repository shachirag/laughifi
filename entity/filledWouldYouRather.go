package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FilledWouldYouRatherEntity struct {
	Id               primitive.ObjectID `json:"id" bson:"_id"`
	UserId           primitive.ObjectID `json:"userId" bson:"userId"`
	WouldYouRatherId primitive.ObjectID `json:"wouldYouRatherId" bson:"wouldYouRatherId"`
	Answer           string             `json:"answer" bson:"answer"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
}
