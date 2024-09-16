package objectStorage

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/api/option"
)

type ManagerT struct {
	Ctx context.Context
	S3  S3T
	GCS GCST
}

type S3T struct {
	Client          *minio.Client
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Secure          bool
}

type GCST struct {
	Client          *storage.Client
	CredentialsFile string
}

type ObjectT struct {
	BucketName string      `json:"bucket"`
	ObjectPath string      `json:"path"`
	Info       ObjectInfoT `json:"-"`
}

type ObjectInfoT struct {
	Exist       bool
	MD5         string
	Size        int64
	ContentType string
}

func NewManager(ctx context.Context, s3 S3T, gcs GCST) (man ManagerT, err error) {
	man.Ctx = ctx

	man.S3 = s3
	man.S3.Client, err = minio.New(
		man.S3.Endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(man.S3.AccessKeyID, man.S3.SecretAccessKey, ""),
			Region: man.S3.Region,
			Secure: man.S3.Secure,
		},
	)
	if err != nil {
		return man, err
	}

	man.GCS.CredentialsFile = gcs.CredentialsFile
	man.GCS.Client, err = storage.NewClient(man.Ctx, option.WithCredentialsFile(man.GCS.CredentialsFile))

	return man, err
}

func (m *ManagerT) S3ObjectExist(obj ObjectT) (info ObjectInfoT, err error) {
	stat, err := m.S3.Client.StatObject(m.Ctx, obj.BucketName, obj.ObjectPath, minio.GetObjectOptions{})
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

func (m *ManagerT) GCSObjectExist(obj ObjectT) (info ObjectInfoT, err error) {
	stat, err := m.GCS.Client.Bucket(obj.BucketName).Object(obj.ObjectPath).Attrs(m.Ctx)
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

func (m *ManagerT) S3DownloadObjectBytes(obj ObjectT) (b []byte, err error) {
	// Descargar el objeto desde el bucket
	object, err := m.S3.Client.GetObject(m.Ctx, obj.BucketName, obj.ObjectPath, minio.GetObjectOptions{})
	if err != nil {
		return b, err
	}
	defer object.Close()

	// read the content from the buffer
	contentBytes := new(bytes.Buffer)
	_, err = io.Copy(contentBytes, object)
	if err != nil {
		return nil, err
	}

	b = contentBytes.Bytes()

	return b, err
}

func (m *ManagerT) TransferObjectFromGCSToS3(src, dst ObjectT) (info ObjectInfoT, err error) {
	srcReader, err := m.GCS.Client.Bucket(src.BucketName).Object(src.ObjectPath).NewReader(m.Ctx)
	if err != nil {
		return info, err
	}
	defer srcReader.Close()

	_, err = m.S3.Client.PutObject(m.Ctx, dst.BucketName, dst.ObjectPath, srcReader, srcReader.Attrs.Size,
		minio.PutObjectOptions{
			CacheControl:    srcReader.Attrs.CacheControl,
			ContentEncoding: srcReader.Attrs.ContentEncoding,
			ContentType:     srcReader.Attrs.ContentType,
		})
	if err != nil {
		return info, err
	}

	info, err = m.S3ObjectExist(dst)

	return info, err
}

func (o *ObjectT) String() string {
	return fmt.Sprintf("{bucket: '%s', object: '%s'}", o.BucketName, o.ObjectPath)
}

func (o *ObjectT) Marshal() (result []byte, err error) {
	result, err = json.Marshal(o)
	return result, err
}
