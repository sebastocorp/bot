package objectStorage

import (
	"bot/api/v1alpha3"
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3ManagerT struct {
	ctx    context.Context
	client *minio.Client
}

type S3ObjectT struct {
	reader      io.ReadCloser
	md5Sum      string
	size        int64
	contentType string
}

func (m *S3ManagerT) Init(ctx context.Context, config v1alpha3.SourceConfigT) (err error) {
	m.ctx = ctx
	m.client, err = minio.New(
		config.S3.Endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(config.S3.AccessKeyID, config.S3.SecretAccessKey, ""),
			Region: config.S3.Region,
			Secure: config.S3.Secure,
		},
	)

	return err
}

func (m *S3ManagerT) GetObject(obj ObjectT) (ro ObjectI, err error) {
	stat, err := m.client.StatObject(m.ctx, obj.Bucket, obj.Path, minio.StatObjectOptions{})
	if err != nil {
		return ro, err
	}

	s3obj, err := m.client.GetObject(m.ctx, obj.Bucket, obj.Path, minio.GetObjectOptions{})
	if err != nil {
		return ro, err
	}

	s3obji := &S3ObjectT{}
	s3obji.reader = s3obj
	s3obji.md5Sum = stat.ETag
	s3obji.size = stat.Size
	s3obji.contentType = stat.ContentType

	ro = s3obji
	return ro, err
}

func (m *S3ManagerT) PutObject(obj ObjectT, ro ObjectI) (err error) {
	_, err = m.client.PutObject(m.ctx, obj.Bucket, obj.Path, ro, ro.GetSize(), minio.PutObjectOptions{
		ContentType: ro.GetContentType(),
	})
	if err != nil {
		return err
	}

	return err
}

func (o *S3ObjectT) GetContentType() string {
	return o.contentType
}

func (o *S3ObjectT) GetSize() int64 {
	return o.size
}

func (o *S3ObjectT) GetMD5String() string {
	return o.md5Sum
}

func (o *S3ObjectT) Read(p []byte) (n int, err error) {
	n, err = o.reader.Read(p)
	return n, err
}

func (o *S3ObjectT) Close() error {
	err := o.reader.Close()
	return err
}
