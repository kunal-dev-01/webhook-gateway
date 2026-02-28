package utils

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	
	"support/internal/config"
)

// ============================================================================
// R2 Client
// ============================================================================

type R2Client struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
	publicURL string
}

func NewR2Client(cfg *config.R2Config) (*R2Client, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, opts ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID),
		}, nil
	})

	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithEndpointResolverWithOptions(r2Resolver),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		awsConfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create R2 config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	return &R2Client{
		client:    s3Client,
		presigner: s3.NewPresignClient(s3Client),
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
	}, nil
}

// ============================================================================
// Upload
// ============================================================================

// Upload uploads a file to R2 and returns the public URL
func (r *R2Client) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	return fmt.Sprintf("%s/%s", r.publicURL, key), nil
}

// Delete deletes an object from R2
func (r *R2Client) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	return err
}

// ============================================================================
// Accessors
// ============================================================================

// PublicURL returns the full public URL for a given key (used for avatars, etc.)
func (r *R2Client) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s", r.publicURL, key)
}

// BucketName returns the configured bucket name (for storing in DB records)
func (r *R2Client) BucketName() string {
	return r.bucket
}

// ============================================================================
// Presigned URLs (for private file downloads like support attachments)
// ============================================================================

// GeneratePresignedURL generates a time-limited presigned GET URL for a private object.
// fileName is used to set Content-Disposition so browser downloads with correct name.
func (r *R2Client) GeneratePresignedURL(ctx context.Context, key, fileName string, ttl time.Duration) (string, error) {
	req, err := r.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(r.bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(fmt.Sprintf(`attachment; filename="%s"`, fileName)),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = ttl
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}