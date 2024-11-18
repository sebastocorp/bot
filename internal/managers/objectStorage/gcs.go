package objectStorage

import (
	"bot/api/v1alpha3"
	"context"
	"encoding/hex"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCSManagerT struct {
	ctx    context.Context
	client *storage.Client
}

type GCSObjectT struct {
	reader      io.ReadCloser
	md5Sum      string
	size        int64
	contentType string
}

func (m *GCSManagerT) Init(ctx context.Context, config v1alpha3.SourceConfigT) (err error) {
	m.ctx = ctx
	m.client, err = storage.NewClient(m.ctx, option.WithCredentialsFile(config.GCS.CredentialsFile))

	return err
}

func (m *GCSManagerT) GetObject(obj ObjectT) (ro ObjectI, err error) {
	objgcs := m.client.Bucket(obj.Bucket).Object(obj.Path)
	stat, err := objgcs.Attrs(m.ctx)
	if err != nil {
		return ro, err
	}

	if len(stat.MD5) == 0 {
		err = fmt.Errorf("object '%s' without md5 assosiated in '%s' source bucket", obj.Path, obj.Bucket)
		return ro, err
	}

	objgcsi := &GCSObjectT{}
	objgcsi.reader, err = objgcs.NewReader(m.ctx)
	if err != nil {
		return ro, err
	}
	objgcsi.md5Sum = hex.EncodeToString(stat.MD5)
	objgcsi.size = stat.Size
	objgcsi.contentType = stat.ContentType

	ro = objgcsi
	return ro, err
}

func (m *GCSManagerT) PutObject(obj ObjectT, ro ObjectI) (err error) {
	gcsobj := m.client.Bucket(obj.Bucket).Object(obj.Path)
	wo := gcsobj.NewWriter(m.ctx)
	defer wo.Close()
	wo.ContentType = ro.GetContentType()
	wo.Size = ro.GetSize()
	wo.MD5, err = hex.DecodeString(ro.GetMD5String())
	if err != nil {
		return err
	}

	_, err = io.Copy(wo, ro)
	return err
}

func (o *GCSObjectT) GetContentType() string {
	return o.contentType
}

func (o *GCSObjectT) GetSize() int64 {
	return o.size
}

func (o *GCSObjectT) GetMD5String() string {
	return o.md5Sum
}

func (o *GCSObjectT) Read(p []byte) (n int, err error) {
	n, err = o.reader.Read(p)
	return n, err
}

func (o *GCSObjectT) Close() error {
	err := o.reader.Close()
	return err
}
