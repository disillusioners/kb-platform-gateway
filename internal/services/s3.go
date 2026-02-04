package services

import (
	"fmt"
	"io"
	"time"

	"kb-platform-gateway/internal/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Client wraps the S3 service with presigned URL generation
type S3Client struct {
	sess   *session.Session
	bucket string
	client *s3.S3
}

// NewS3Client creates a new S3 client
func NewS3Client(cfg *config.S3Config) (*S3Client, error) {
	var creds *credentials.Credentials
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		creds = credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, "")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.Region),
		Endpoint:    aws.String(cfg.Endpoint),
		Credentials: creds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 session: %w", err)
	}

	return &S3Client{
		sess:   sess,
		bucket: cfg.Bucket,
		client: s3.New(sess),
	}, nil
}

// GenerateUploadPresignedURL generates a presigned URL for uploading an object
func (c *S3Client) GenerateUploadPresignedURL(key string, contentType string, expiresIn time.Duration) (string, error) {
	req, _ := c.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})

	url, err := req.Presign(expiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

// GenerateDownloadPresignedURL generates a presigned URL for downloading an object
func (c *S3Client) GenerateDownloadPresignedURL(key string, expiresIn time.Duration) (string, error) {
	req, _ := c.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(expiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

// UploadDocument uploads a document from a reader to S3
func (c *S3Client) UploadDocument(key string, contentType string, body io.Reader) error {
	uploader := s3manager.NewUploader(c.sess)

	input := &s3manager.UploadInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        body,
	}

	_, err := uploader.Upload(input)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// DownloadDocument downloads a document from S3 to a writer
func (c *S3Client) DownloadDocument(key string, writer io.WriterAt) (int64, error) {
	downloader := s3manager.NewDownloader(c.sess)

	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	n, err := downloader.Download(writer, input)
	if err != nil {
		return 0, fmt.Errorf("failed to download from S3: %w", err)
	}

	return n, nil
}

// DeleteDocument deletes a document from S3
func (c *S3Client) DeleteDocument(key string) error {
	_, err := c.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// DocumentExists checks if a document exists in S3
func (c *S3Client) DocumentExists(key string) (bool, error) {
	_, err := c.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}
