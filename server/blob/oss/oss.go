package oss

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Service struct {
	client     *oss.Client
	ossId      string
	ossSecrets string
}

const (
	bucketName = "clive-coolcar"
)

func NewService(addr string, ossId string, ossSecrets string) (*Service, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse addr: %v", err)
	}

	client, err := oss.New(u.String(), ossId, ossSecrets)
	if err != nil {
		return nil, fmt.Errorf("cannot create OSS client: %v", err)
	}

	return &Service{
		client:     client,
		ossId:      ossId,
		ossSecrets: ossSecrets,
	}, nil
}

func (s *Service) SignUrl(ctx context.Context, httpMethod oss.HTTPMethod, path string, timeout time.Duration) (string, error) {
	bucket, err := s.client.Bucket(bucketName)
	if err != nil {
		return "", err
	}

	var opts []oss.Option
	if httpMethod == oss.HTTPPut {
		opts = append(opts, oss.ContentType("application/octet-stream"))
	}

	u, err := bucket.SignURL(path, httpMethod, int64(timeout.Seconds()), opts...)
	if err != nil {
		return "", err
	}

	return u, nil
}
func (s *Service) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	bucket, err := s.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return bucket.GetObject(path)
}
