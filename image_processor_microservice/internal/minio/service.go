package minio

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"os"
	"strconv"
)

const (
	BucketSourceName = "image-source"
	BucketTargetName = "image-result"
)

type Client struct {
	ctx         context.Context
	minioClient *minio.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	useSsl, err := strconv.ParseBool(os.Getenv("minio_use_ssl"))
	if err != nil {
		return nil, err
	}
	minioClient, err := minio.New(os.Getenv("minio_address"), &minio.Options{
		Creds: credentials.NewStaticV4(
			os.Getenv("minio_access_key_id"),
			os.Getenv("minio_secret_access_key"),
			os.Getenv("minio_token"),
		),
		Secure: useSsl,
	})
	if err != nil {
		return nil, err
	}

	cli := &Client{ctx: ctx, minioClient: minioClient}

	err = cli.bucketInit()
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func (c *Client) Get(bucketName string, filename string) ([]byte, error) {
	object, err := c.minioClient.GetObject(c.ctx, bucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	result, err := io.ReadAll(object)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Upsert(content []byte, filename string) error {
	reader := bytes.NewReader(content)
	_, err := c.minioClient.PutObject(c.ctx, BucketTargetName, filename, reader, int64(len(content)), minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) bucketInit() error {
	exists, err := c.minioClient.BucketExists(c.ctx, BucketTargetName)
	if err != nil {
		return err
	}

	if !exists {
		err = c.minioClient.MakeBucket(c.ctx, BucketTargetName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
