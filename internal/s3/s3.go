package internal_s3

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3Enabled = false
var S3client *s3.Client
var AllowedBuckets []string

func Connect() {
	if os.Getenv("PDFIFY_USE_S3") != "true" {
		log.Println("PDFIFY_USE_S3 does not equal true, will return PDFs directly.")
		return
	}

	if os.Getenv("PDFIFY_ALLOWED_BUCKETS") == "" {
		log.Panicf("PDFIFY_ALLOWED_BUCKETS must be set to a list of buckets e.g: test_bucket,test_bucket_2")
		return
	}

	if os.Getenv("PDFIFY_AWS_REGION") == "" {
		log.Panicf("PDFIFY_AWS_REGION must be set to region of S3 buckets")
		return
	}

	AllowedBuckets = strings.Split(os.Getenv("PDFIFY_ALLOWED_BUCKETS"), ",")

	// Load the default AWS configuration (can include region, credentials, etc.)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("PDFIFY_AWS_REGION")))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	// Create S3 client
	S3client = s3.NewFromConfig(cfg)

	S3Enabled = true
}
