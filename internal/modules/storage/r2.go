package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/delordemm1/qplayground/internal/platform"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewR2Storage() (*R2Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			platform.ENV_R2_ACCESS_KEY_ID,
			platform.ENV_R2_SECRET_ACCESS_KEY,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", platform.ENV_CLOUDFLARE_ACCOUNT_ID))
	})

	return &R2Storage{
		client:    client,
		bucket:    platform.ENV_R2_BUCKET_NAME,
		publicURL: platform.ENV_R2_PUBLIC_URL,
	}, nil
}

func (r *R2Storage) Upload(ctx context.Context, key string, data io.Reader, options *UploadOptions) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
		Body:   data,
	}

	if options != nil {
		if options.ContentType != "" {
			input.ContentType = aws.String(options.ContentType)
		}
		if len(options.Metadata) > 0 {
			input.Metadata = options.Metadata
		}
	}

	_, err := r.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload object to R2: %w", err)
	}

	slog.Info("Successfully uploaded object to R2", "key", key)
	return nil
}

func (r *R2Storage) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from R2: %w", err)
	}

	slog.Info("Successfully deleted object from R2", "key", key)
	return nil
}

func (r *R2Storage) GetPublicURL(key string) string {
	return fmt.Sprintf("%s/%s", r.publicURL, key)
}
