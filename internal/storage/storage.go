package storage

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type BucketBasics struct {
	S3Client *s3.Client
}

var (
	once        sync.Once
	s3Singleton *BucketBasics
)

func LoadStorage() *BucketBasics {
	once.Do(func() {
		log.Printf("Initializing S3 storage client")
		s3Singleton = initS3()
		log.Printf("S3 storage client initialized successfully")

	})
	log.Printf("You use S3 storage client")

	return s3Singleton
}

func initS3() *BucketBasics {

	endpoint := "https://storage.yandexcloud.net"
	region := "ru-central1"

	accessKey := os.Getenv("YANDEX_ACCESS_KEY")
	secretKey := os.Getenv("YANDEX_SECRET_KEY")
	_ = os.Getenv("YANDEX_BUCKET")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		})),
	)

	if err != nil {
		log.Fatalf("ошибка инициализации AWS SDK: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	return &BucketBasics{
		S3Client: s3Client}
}

func (basics *BucketBasics) UploadFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, bucketName string, objectKey string, fileName string) error {
	_, err := basics.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        file,
		ACL:         "public-read", // если нужно сразу доступно извне
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			log.Printf("Error while uploading object to %s. The object is too large.\n"+
				"To upload objects larger than 5GB, use the S3 console (160GB max)\n"+
				"or the multipart upload API (5TB max).", bucketName)
		} else {
			log.Printf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
				fileName, bucketName, objectKey, err)
		}
	} else {
		err = s3.NewObjectExistsWaiter(basics.S3Client).Wait(
			ctx, &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: aws.String(objectKey)}, time.Minute)
		if err != nil {
			log.Printf("Failed attempt to wait for object %s to exist.\n", objectKey)
		}
	}
	return err
}
