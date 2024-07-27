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
	"go.mongodb.org/mongo-driver/mongo"
)

func AdminForgotPassword(ctx context.Context, db *database.DB, sesClient *ses.Client, input model.ForgotPasswordRequestInput) (*model.Response, error) {
	var (
		adminColl = db.GetCollection("admin")
		otpColl   = db.GetCollection("otp")
		admin     entity.AdminEntity
	)

	smallEmail := strings.ToLower(input.Email)

	filter := bson.M{"email": smallEmail}

	err := adminColl.FindOne(ctx, filter).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("admin not found")
		}
		return nil, gqlerror.Errorf("Internal server error while fetching the admin.")
	}

	otp := utils.Generate6DigitOtp()
	// otp := "111111"

	otpData := entity.OtpEntity{
		Id:        primitive.NewObjectID(),
		Otp:       otp,
		Email:     smallEmail,
		CreatedAt: time.Now().UTC(),
	}

	_, err = otpColl.InsertOne(ctx, otpData)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to store OTP in the database")
	}

	err = utils.SendForgotPasswordEmail(admin.Email, admin.FirstName+" "+admin.LastName, otp)
	if err != nil {
		return nil, gqlerror.Errorf("Internal server error while sending the email" + err.Error())
	}

	return &model.Response{
		Message: "Otp Sent Successfully",
	}, nil
}
