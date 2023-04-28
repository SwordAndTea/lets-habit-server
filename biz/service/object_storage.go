package service

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/swordandtea/lets-habit-server/biz/config"
	"io"
	"time"
)

type ObjectStorage interface {
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
	PutObject(ctx context.Context, key string, data io.Reader) error
	ObjectKeyToURL(key string, ttl time.Duration) (string, error)
}

var currentObjectStorageImpl ObjectStorage

func GetObjectStorageExecutor() ObjectStorage {
	return currentObjectStorageImpl
}

type ObjectStorageImplS3 struct {
	bucket string
	cli    *s3.S3
}

func (impl *ObjectStorageImplS3) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := impl.cli.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(impl.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return out.Body, nil
}

func (impl *ObjectStorageImplS3) PutObject(ctx context.Context, key string, data io.Reader) error {
	_, err := impl.cli.PutObject(&s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(data),
		Bucket: aws.String(impl.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}

func (impl *ObjectStorageImplS3) ObjectKeyToURL(key string, ttl time.Duration) (string, error) {
	req, _ := impl.cli.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(impl.bucket),
		Key:    aws.String(key),
	})
	signedURL, err := req.Presign(ttl)
	if err != nil {
		return "", err
	}
	return signedURL, nil
}

func InitObjectStorage(conf *config.ObjectStorageConfig) error {
	s, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(conf.AK, conf.SK, ""),
		Endpoint:         aws.String(conf.Endpoint),
		Region:           aws.String(conf.Region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return err
	}
	s3Cli := s3.New(s)
	currentObjectStorageImpl = &ObjectStorageImplS3{
		bucket: conf.Bucket,
		cli:    s3Cli,
	}
	return nil
}
