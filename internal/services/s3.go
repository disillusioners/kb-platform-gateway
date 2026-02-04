package services

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"kb-platform-gateway/internal/config"
)

type S3Client struct {
	client *s3.Client
	cfg    *config.S3Config
}

func NewS3Client(cfg *config.S3Config) (*S3Client, error) {
	cfgAWS, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     cfg.AccessKeyID,
				SecretAccessKey: cfg.SecretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		return nil, err
	}

	clientOptions := []func(*s3.Options){}
	if cfg.Endpoint != "" {
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.BaseEndpoint = &cfg.Endpoint
		})
	}

	client := s3.NewFromConfig(cfgAWS, clientOptions...)

	return &S3Client{
		client: client,
		cfg:    cfg,
	}, nil
}

func (c *S3Client) GeneratePresignedUploadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.client)

	presignResult, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &c.cfg.Bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expires))

	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
}

func (c *S3Client) GeneratePresignedDownloadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.client)

	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &c.cfg.Bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expires))

	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
}

func (c *S3Client) DeleteObject(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &c.cfg.Bucket,
		Key:    &key,
	})
	return err
}
