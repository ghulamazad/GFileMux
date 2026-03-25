package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ghulamazad/GFileMux"
	"github.com/ghulamazad/GFileMux/utils"
)

// S3Options holds configuration options for interacting with an S3 store.
type S3Options struct {
	DebugMode    bool
	UsePathStyle bool
	ACL          types.ObjectCannedACL
}

// S3Store is a structure that represents the S3 storage client.
type S3Store struct {
	client  *s3.Client
	options S3Options
}

// NewS3FromConfig initializes an S3Store using an AWS configuration.
func NewS3FromConfig(cfg aws.Config, options S3Options) (*S3Store, error) {
	client := s3.NewFromConfig(cfg, func(opt *s3.Options) {
		opt.UsePathStyle = options.UsePathStyle
		if options.DebugMode {
			opt.ClientLogMode = aws.LogSigning | aws.LogRequest | aws.LogResponseWithBody
		}
	})
	return &S3Store{client, options}, nil
}

// NewS3FromEnvironment initializes an S3Store from the environment configuration.
func NewS3FromEnvironment(options S3Options) (*S3Store, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return NewS3FromConfig(cfg, options)
}

// NewS3FromClient initializes an S3Store from an existing S3 client.
func NewS3FromClient(client *s3.Client, options S3Options) (*S3Store, error) {
	return &S3Store{client, options}, nil
}

// Upload uploads a file to S3 with the given options.
func (s *S3Store) Upload(ctx context.Context, r io.Reader, options *GFileMux.UploadFileOptions) (*GFileMux.UploadedFileMetadata, error) {
	if options == nil {
		return nil, errors.New("upload options are required")
	}
	if len(strings.TrimSpace(options.Bucket)) == 0 {
		return nil, errors.New("please provide a valid S3 bucket")
	}

	// Buffer the reader so we can compute the size and seek back for upload.
	b := new(bytes.Buffer)
	r = io.TeeReader(r, b)
	n, err := io.Copy(io.Discard, r)
	if err != nil {
		return nil, err
	}

	seeker, err := utils.ReaderToSeeker(b)
	if err != nil {
		return nil, err
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:   aws.String(options.Bucket),
		Metadata: options.Metadata,
		Key:      aws.String(options.FileName),
		ACL:      s.options.ACL,
		Body:     seeker,
	})
	if err != nil {
		return nil, &GFileMux.StorageError{Backend: "s3", Op: "Upload", Err: err}
	}

	return &GFileMux.UploadedFileMetadata{
		FolderDestination: options.Bucket,
		Size:              n,
		Key:               options.FileName,
	}, nil
}

// Path generates a URL to access a file in S3, either a presigned URL or a direct URL.
func (s *S3Store) Path(ctx context.Context, options GFileMux.PathOptions) (string, error) {
	if !options.IsSecure {
		resp, err := s.client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
			Bucket: &options.Bucket,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get bucket location: %w", err)
		}

		region := string(resp.LocationConstraint)
		if region == "" {
			region = "us-east-1"
		}
		url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", options.Bucket, region, options.Key)
		return url, nil
	}

	presignClient := s3.NewPresignClient(s.client)
	presignRequest, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &options.Bucket,
		Key:    &options.Key,
	}, s3.WithPresignExpires(options.ExpirationTime))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignRequest.URL, nil
}

// Delete removes an object from S3 identified by bucket and key.
func (s *S3Store) Delete(ctx context.Context, bucket, key string) error {
	if bucket == "" || key == "" {
		return fmt.Errorf("bucket and key are required")
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return &GFileMux.StorageError{Backend: "s3", Op: "Delete", Err: err}
	}
	return nil
}

// Close closes the S3 store (no-op; AWS SDK manages its own connections).
func (s *S3Store) Close() error {
	if s.options.DebugMode {
		log.Println("S3 store is being closed.")
	}
	return nil
}
