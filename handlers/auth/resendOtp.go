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
	"go.mongodb.org/mongo-driver/mongo"
)

func ResendOtp(ctx context.Context, db *database.DB, sesClient *ses.Client, input model.ResendOtpRequestInput) (*model.Response, error) {
	var (
		otpColl = db.GetCollection("otp")
		otpData entity.OtpEntity
	)

	smallEmail := strings.ToLower(input.Email)

	err := otpColl.FindOne(ctx, bson.M{"email": smallEmail}).Decode(&otpData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("No Otp Found")
		}
		return nil, gqlerror.Errorf("Internal server error while fetching OTP data: " + err.Error())

	}

	newOTP := utils.Generate6DigitOtp()

	otpData.Otp = newOTP
	otpData.CreatedAt = time.Now().UTC()

	update := bson.M{"$set": bson.M{"otp": newOTP, "createdAt": otpData.CreatedAt}}

	_, err = otpColl.UpdateOne(ctx, bson.M{"email": smallEmail}, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update Otp")
	}

	var customerData entity.CustomerEntity
	err = db.GetCollection("customer").FindOne(ctx, bson.M{"email": smallEmail}).Decode(&customerData)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to fetch user")
	}

	_, err = utils.SendEmail(sesClient, customerData.Name, newOTP)
	if err != nil {
		return nil, gqlerror.Errorf("Internal server error while sending the email")
	}

	return &model.Response{
		Message: "Otp Resent Successfully",
	}, nil

}
