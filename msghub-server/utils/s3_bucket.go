package utils

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/joho/godotenv"
	"io"
	"log"
	"os"
)

var (
	AWS_S3_REGION string
	AWS_S3_BUCKET string
)

func StoreThisFileInBucket(folderName, filename string, file io.Reader) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(err.Error())
		os.Exit(0)
	}

	AWS_S3_REGION = os.Getenv("AWS_S3_REGION")
	AWS_S3_BUCKET = os.Getenv("AWS_S3_BUCKET")

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(AWS_S3_REGION),
		},
	)
	if err != nil {
		log.Println(err)
		return ""
	}

	uploader := s3manager.NewUploader(sess)
	fmt.Println("We here")

	res, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AWS_S3_BUCKET),         // Bucket to be used
		Key:    aws.String(folderName + filename), // Name of the file to be saved
		Body:   file,                              // File
	})
	if err != nil {
		// Do your error handling here
		fmt.Println(err)
		return ""
	}

	return res.Location
}
