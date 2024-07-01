package auth

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"laughifi/utils"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Signup(ctx context.Context, db *database.DB, sesClient *ses.Client, data model.SignUpRequestInput) (*model.Response, error) {
	var (
		otpColl      = db.GetCollection("otp")
		customerColl = db.GetCollection("customer")
	)

	smallEmail := strings.ToLower(data.Email)

	filter := bson.M{
		"email": smallEmail,
	}

	exists, err := customerColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Database error: " + err.Error())
	}

	if exists > 0 {
		return nil, gqlerror.Errorf("Email is already in use.")
	}

	id := primitive.NewObjectID()

	otp := utils.Generate6DigitOtp()
	otpData := entity.OtpEntity{
		Id:        id,
		Otp:       otp,
		Email:     data.Email,
		CreatedAt: time.Now().UTC(),
	}

	_, err = otpColl.InsertOne(ctx, otpData)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to insert otp")
	}

	_, err = utils.SendEmail(sesClient, data.Email, otp)
	if err != nil {
		return nil, gqlerror.Errorf("Error sending OTP to email: " + err.Error())
	}

	return &model.Response{
		Message: "Otp Sent Successfully",
	}, nil
}
