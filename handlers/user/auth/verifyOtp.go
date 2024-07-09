package auth

import (
	"context"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func VerifyOtpForResetPassword(ctx context.Context, db *database.DB, input model.VerifyOtpForResetPasswordRequestInput) (*model.Response, error) {
	var (
		otpColl = db.GetCollection("otp")
		otpData entity.OtpEntity
		user    entity.CustomerEntity
	)

	if input.Otp == "" {
		return nil, gqlerror.Errorf("OTP is required")
	}

	smallEmail := strings.ToLower(input.Email)

	err := db.GetCollection("customer").FindOne(ctx, bson.M{"email": smallEmail}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("user not found")
		}
		return nil, gqlerror.Errorf("Internal server error while fetching the user.")
	}

	err = otpColl.FindOne(ctx, bson.M{"email": smallEmail}, options.FindOne().SetSort(bson.M{"createdAt": -1})).Decode(&otpData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("Invalid OTP")
		}
		return nil, gqlerror.Errorf("Internal server error while fetching OTP: " + err.Error())
	}

	timeDiff := time.Since(otpData.CreatedAt)
	if timeDiff.Minutes() > 10 {
		return nil, gqlerror.Errorf("OTP expired, generate a new OTP.")
	}

	if input.Otp != otpData.Otp {
		return nil, gqlerror.Errorf("OTP does not match")
	}

	return &model.Response{
		Message: "Otp Verified Successfully",
	}, nil
}
