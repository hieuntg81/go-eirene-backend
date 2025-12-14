package storage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appconfig "bamboo-rescue/internal/config"
	"go.uber.org/zap"
)

// Client defines the interface for storage operations
type Client interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
	GetURL(key string) string
}

type s3Client struct {
	client    *s3.Client
	bucket    string
	publicURL string
	log       *zap.Logger
}

// NewS3Client creates a new S3-compatible storage client
func NewS3Client(cfg *appconfig.Config, log *zap.Logger) (Client, error) {
	if cfg.S3.Endpoint == "" {
		log.Warn("S3 endpoint not configured, using mock storage")
		return &mockClient{log: log}, nil
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: cfg.S3.Endpoint,
		}, nil
	})

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3.AccessKey,
			cfg.S3.SecretKey,
			"",
		)),
		config.WithRegion(cfg.S3.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	publicURL := cfg.S3.PublicURL
	if publicURL == "" {
		publicURL = cfg.S3.Endpoint
	}

	return &s3Client{
		client:    client,
		bucket:    cfg.S3.Bucket,
		publicURL: publicURL,
		log:       log,
	}, nil
}

func (c *s3Client) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	_, err := c.client.PutObject(ctx, input)
	if err != nil {
		c.log.Error("Failed to upload to S3", zap.String("key", key), zap.Error(err))
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return c.GetURL(key), nil
}

func (c *s3Client) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	_, err := c.client.DeleteObject(ctx, input)
	if err != nil {
		c.log.Error("Failed to delete from S3", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (c *s3Client) GetURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", c.publicURL, c.bucket, key)
}

// Mock client for development
type mockClient struct {
	log *zap.Logger
}

func (c *mockClient) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	c.log.Debug("Mock upload", zap.String("key", key), zap.Int("size", len(data)))
	return fmt.Sprintf("http://localhost:8080/mock-storage/%s", key), nil
}

func (c *mockClient) Delete(ctx context.Context, key string) error {
	c.log.Debug("Mock delete", zap.String("key", key))
	return nil
}

func (c *mockClient) GetURL(key string) string {
	return fmt.Sprintf("http://localhost:8080/mock-storage/%s", key)
}
