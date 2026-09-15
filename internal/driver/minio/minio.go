package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/yourname/mcp-x/internal/driver"
)

type MinIODriver struct {
	client   *minio.Client
	endpoint string
}

func (d *MinIODriver) Name() string            { return "minio" }
func (d *MinIODriver) Type() driver.DriverType { return driver.DriverTypeNoSQL }

func (d *MinIODriver) Connect(ctx context.Context, cfg driver.ConnConfig) error {
	endpoint := cfg.Endpoint
	if endpoint == "" && len(cfg.Endpoints) > 0 {
		endpoint = cfg.Endpoints[0]
	}
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return fmt.Errorf("minio new: %w", err)
	}
	if _, err := cli.ListBuckets(ctx); err != nil {
		return fmt.Errorf("minio ping: %w", err)
	}
	d.client = cli
	d.endpoint = endpoint
	return nil
}

func (d *MinIODriver) ListBuckets(ctx context.Context) ([]driver.BucketInfo, error) {
	buckets, err := d.client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("minio list buckets: %w", err)
	}
	result := make([]driver.BucketInfo, 0, len(buckets))
	for _, b := range buckets {
		result = append(result, driver.BucketInfo{
			Name:        b.Name,
			ObjectCount: -1,
		})
	}
	return result, nil
}

func (d *MinIODriver) ListObjects(ctx context.Context, bucket, prefix string, limit int) ([]driver.ObjectInfo, error) {
	if limit <= 0 {
		limit = 100
	}
	if prefix == "" {
		prefix = ""
	}
	objCh := d.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})
	var result []driver.ObjectInfo
	for obj := range objCh {
		if obj.Err != nil {
			return nil, obj.Err
		}
		result = append(result, driver.ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			ContentType:  obj.ContentType,
			LastModified: obj.LastModified.Format("2006-01-02T15:04:05Z"),
		})
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (d *MinIODriver) GetObject(ctx context.Context, bucket, key string) (string, error) {
	obj, err := d.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("minio get: %w", err)
	}
	defer obj.Close()
	data, err := io.ReadAll(io.LimitReader(obj, maxObjectBytes))
	if err != nil {
		return "", fmt.Errorf("minio read: %w", err)
	}
	return string(data), nil
}

const maxObjectBytes = 1 << 20 // 1MB ponytail: raise when larger reads needed

func (d *MinIODriver) PutObject(ctx context.Context, bucket, key, contentType string, body []byte) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := d.client.PutObject(ctx, bucket, key, bytes.NewReader(body), int64(len(body)),
		minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (d *MinIODriver) DeleteObject(ctx context.Context, bucket, key string) error {
	err := d.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	return err
}

func (d *MinIODriver) Ping(ctx context.Context) error {
	_, err := d.client.ListBuckets(ctx)
	return err
}

func (d *MinIODriver) Close() error {
	return nil
}

func init() {
	driver.Register("minio", func() driver.AnyDriver { return &MinIODriver{} })
}
