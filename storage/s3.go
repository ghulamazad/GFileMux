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

	return &S3Store{
		client,
		options,
	}, nil
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
	return &S3Store{
		client,
		options,
	}, nil
}

// Upload uploads a file to S3 with the given options.
func (s *S3Store) Upload(ctx context.Context, r io.Reader, options GFileMux.UploadFileOptions) (*GFileMux.UploadedFileMetadata, error) {
	// Ensure the S3 bucket is valid
	if len(strings.TrimSpace(options.Bucket)) == 0 {
		return nil, errors.New("please provide a valid S3 bucket")
	}

	// Create a buffer to store the contents of the file
	b := new(bytes.Buffer)
	r = io.TeeReader(r, b)

	// Copy the content to discard to calculate the size
	n, err := io.Copy(io.Discard, r)
	if err != nil {
		return nil, err
	}

	// Convert the buffer to a reader that can be seeked
	seeker, err := utils.ReaderToSeeker(b)
	if err != nil {
		return nil, err
	}

	// Upload the file to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:   aws.String(options.Bucket),
		Metadata: options.Metadata,
		Key:      aws.String(options.FileName),
		ACL:      s.options.ACL,
		Body:     seeker,
	})
	if err != nil {
		return nil, err
	}

	// Return metadata of the uploaded file
	return &GFileMux.UploadedFileMetadata{
		FolderDestination: options.Bucket,
		Size:              n,
		Key:               options.FileName,
	}, nil
}

// Path generates a URL to access a file in S3, either a presigned URL or a direct URL.
func (s *S3Store) Path(ctx context.Context, options GFileMux.PathOptions) (string, error) {
	// If the file should be accessed over HTTP (non-secure), construct a direct URL
	if !options.IsSecure {
		resp, err := s.client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
			Bucket: &options.Bucket,
		})

		if err != nil {
			return "", fmt.Errorf("failed to get bucket location: %w", err)
		}

		// Default to "us-east-1" if no location is provided
		region := string(resp.LocationConstraint)
		if region == "" {
			region = "us-east-1"
		}

		// Construct a direct URL to the S3 object
		url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", options.Bucket, region, options.Key)
		return url, nil
	}

	// Otherwise, create a presigned URL for secure access
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

// Close closes the S3 store (no-op for the AWS SDK but provides an interface for potential cleanup).
func (s *S3Store) Close() error {
	// If DebugMode is enabled, log that the store is being closed.
	if s.options.DebugMode {
		log.Println("S3 store is being closed.")
	}

	// Since AWS SDK for Go doesn't need to explicitly close resources,
	// we will just return nil here.
	return nil
}
