package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	//"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Service struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	region        string
}

// NewS3Service creates a new AWS S3 service
func NewS3Service(ctx context.Context, region, bucket, accessKeyID, secretAccessKey string) (*S3Service, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	return &S3Service{
		client:        client,
		presignClient: presignClient,
		bucket:        bucket,
		region:        region,
	}, nil
}

// ImageFile represents an image file in S3
type ImageFile struct {
	Key        string
	Name       string
	URL        string
	UploadedAt time.Time
}

// ListImages retrieves all images from the S3 bucket
func (s *S3Service) ListImages(ctx context.Context) ([]*ImageFile, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	}

	result, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to list objects: %w", err)
	}

	images := make([]*ImageFile, 0, len(result.Contents))
	for _, obj := range result.Contents {
		if obj.Key == nil {
			continue
		}

		url, err := s.GetImageURL(ctx, *obj.Key)
		if err != nil {
			return nil, fmt.Errorf("unable to get presigned URL for %s: %w", *obj.Key, err)
		}

		images = append(images, &ImageFile{
			Key:        *obj.Key,
			Name:       *obj.Key,
			URL:        url,
			UploadedAt: *obj.LastModified,
		})
	}

	return images, nil
}

// UploadImage uploads an image to S3
func (s *S3Service) UploadImage(ctx context.Context, key string, mimeType string, content io.Reader) (*ImageFile, error) {
	// Read content into buffer
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, content); err != nil {
		return nil, fmt.Errorf("unable to read content: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(mimeType),
		//ACL:         types.ObjectCannedACLPublicRead,
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to upload file: %w", err)
	}

	uploadedAt := time.Now()

	log.Printf("Successfully uploaded %s to S3 bucket %s", key, s.bucket)

	url, err := s.GetImageURL(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get presigned URL for uploaded file: %w", err)
	}

	return &ImageFile{
		Key:        key,
		Name:       key,
		URL:        url,
		UploadedAt: uploadedAt,
	}, nil
}

// GetImageURL returns a presigned URL to view the image with a 15-minute expiration
func (s *S3Service) GetImageURL(ctx context.Context, key string) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		return "", fmt.Errorf("unable to generate presigned URL: %w", err)
	}

	return result.URL, nil
}
