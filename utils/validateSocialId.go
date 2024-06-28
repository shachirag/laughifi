package utils

import (
	"errors"
	"laughifi/graph/model"
)

func ValidateSocialId(data *model.SocialLoginRequestInput) error {

	if StringIsEmpty(data.SocialID) {
		return errors.New("socialId cannot be empty")
	}

	return nil
}
