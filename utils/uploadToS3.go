package utils

import (
	"context"
	"io"
	"laughifi/database"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadToS3(fName string, file io.ReadSeeker) (string, error) {
	s3BucketName := os.Getenv("S3_BUCKET_NAME")
	uploader := database.GetS3Uploader()

	u, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(fName),
		Body:   file,
		ACL:    "public-read",
	})
	if err != nil {
		return "", err
	}

	return u.Location, nil
}
