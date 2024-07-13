package graph

import (
	"laughifi/database"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type Resolver struct {
	DB        *database.DB
	S3Client  *manager.Uploader
	SESClient *ses.Client
	ApiClient *apigatewaymanagementapi.Client
}
