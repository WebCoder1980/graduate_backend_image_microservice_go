package minio

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"os"
)

type Client struct {
	ctx         context.Context
	minioClient *minio.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	minioClient, err := minio.New(os.Getenv("minio_address"), &minio.Options{
		Creds: credentials.NewStaticV4(
			os.Getenv("minio_access_key_id"),
			os.Getenv("minio_secret_access_key"),
			os.Getenv("minio_token"),
		),
		Secure: UseSSL,
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

func (c *Client) Upsert(filename string, content []byte) {
	ctx := context.Background()

	reader := bytes.NewReader(content)
	_, err := c.minioClient.PutObject(ctx, BucketName, filename, reader, int64(len(content)), minio.PutObjectOptions{})
	if err != nil {
		log.Panic(err)
	}
}

func (c *Client) bucketInit() error {
	exists, err := c.minioClient.BucketExists(c.ctx, BucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = c.minioClient.MakeBucket(c.ctx, BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
