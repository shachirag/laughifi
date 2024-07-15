package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TriviaEntity struct {
	Id                 primitive.ObjectID `json:"id" bson:"_id"`
	Question           string             `json:"question" bson:"question"`
	OptionOne          string             `json:"optionOne" bson:"optionOne"`
	OptionTwo          string             `json:"optionTwo" bson:"optionTwo"`
	OptionThree        string             `json:"optionThree" bson:"optionThree"`
	OptionFour         string             `json:"optionFour" bson:"optionFour"`
	CorrectAnswer      string             `json:"correctAnswer" bson:"correctAnswer"`
	DareForWrongAnswer string             `json:"dareForWrongAnswer" bson:"dareForWrongAnswer"`
	IsDeleted          bool               `json:"isDeleted" bson:"isDeleted"`
	CreatedAt          time.Time          `json:"cresatedAt" bson:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt" bson:"updatedAt"`
}
