package objectStorage

import (
	"bot/api/v1alpha1"
	"context"
	"encoding/hex"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/api/option"
)

type ManagerT struct {
	Ctx       context.Context
	S3Client  *minio.Client
	GCSClient *storage.Client
}

func NewManager(ctx context.Context, s3 v1alpha1.S3T, gcs v1alpha1.GCST) (man ManagerT, err error) {
	man.Ctx = ctx

	man.S3Client, err = minio.New(
		s3.Endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(s3.AccessKeyID, s3.SecretAccessKey, ""),
			Region: s3.Region,
			Secure: s3.Secure,
		},
	)
	if err != nil {
		return man, err
	}

	man.GCSClient, err = storage.NewClient(man.Ctx, option.WithCredentialsFile(gcs.CredentialsFile))

	return man, err
}

func (m *ManagerT) S3ObjectExist(obj v1alpha1.ObjectT) (info v1alpha1.ObjectInfoT, err error) {
	stat, err := m.S3Client.StatObject(m.Ctx, obj.BucketName, obj.ObjectPath, minio.GetObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			err = nil
			info.Exist = false
		}

		return info, err
	}

	info.Exist = true
	info.MD5 = stat.ETag
	info.Size = stat.Size
	info.ContentType = stat.ContentType

	return info, err
}

func (m *ManagerT) GCSObjectExist(obj v1alpha1.ObjectT) (info v1alpha1.ObjectInfoT, err error) {
	stat, err := m.GCSClient.Bucket(obj.BucketName).Object(obj.ObjectPath).Attrs(m.Ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			err = nil
			info.Exist = false
		}

		return info, err
	}

	if len(stat.MD5) == 0 {
		err = fmt.Errorf("object '%s' without md5 assosiated in '%s' source bucket", obj.ObjectPath, obj.BucketName)
		return info, err
	}

	info.Exist = true
	info.MD5 = hex.EncodeToString(stat.MD5)
	info.Size = stat.Size
	info.ContentType = stat.ContentType

	return info, err
}

func (m *ManagerT) TransferObjectFromGCSToS3(src, dst v1alpha1.ObjectT) (info v1alpha1.ObjectInfoT, err error) {
	object := m.GCSClient.Bucket(src.BucketName).Object(src.ObjectPath)
	stat, err := object.Attrs(m.Ctx)
	if err != nil {
		return info, err
	}

	info.Exist = true
	info.MD5 = hex.EncodeToString(stat.MD5)
	info.Size = stat.Size
	info.ContentType = stat.ContentType

	srcReader, err := object.NewReader(m.Ctx)
	if err != nil {
		return info, err
	}
	defer srcReader.Close()

	_, err = m.S3Client.PutObject(m.Ctx, dst.BucketName, dst.ObjectPath, srcReader, srcReader.Attrs.Size,
		minio.PutObjectOptions{
			CacheControl:    srcReader.Attrs.CacheControl,
			ContentEncoding: srcReader.Attrs.ContentEncoding,
			ContentType:     srcReader.Attrs.ContentType,
		},
	)

	return info, err
}
