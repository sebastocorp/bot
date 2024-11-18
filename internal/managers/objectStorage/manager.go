package objectStorage

import (
	"bot/api/v1alpha3"
	"context"
	"fmt"
	"io"
)

type ObjectManagerI interface {
	Init(ctx context.Context, config v1alpha3.SourceConfigT) error
	GetObject(obj ObjectT) (obji ObjectI, err error)
	PutObject(obj ObjectT, ro ObjectI) (err error)
}

type ObjectI interface {
	io.ReadCloser
	GetContentType() string
	GetSize() int64
	GetMD5String() string
}

type ObjectT struct {
	Bucket string `json:"bucket"`
	Path   string `json:"path"`
}

func GetManager(ctx context.Context, config v1alpha3.SourceConfigT) (m ObjectManagerI, err error) {
	switch config.Type {
	case "s3":
		{
			m = &S3ManagerT{}
		}
	case "gcs":
		{
			m = &GCSManagerT{}
		}
	}
	err = m.Init(ctx, config)
	return m, err
}

func (o *ObjectT) String() string {
	return fmt.Sprintf("{bucket: '%s', object: '%s'}", o.Bucket, o.Path)
}
