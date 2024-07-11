package graph

import (
	"laughifi/database"
	"laughifi/utils/subscription"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type Resolver struct {
	DB              *database.DB
	S3Client        *manager.Uploader
	SESClient       *ses.Client
	SubscriptionMgr *subscription.Manager
}
