package v1alpha1

import "fmt"

type ObjectT struct {
	Bucket string      `json:"bucket"`
	Path   string      `json:"path"`
	Info   ObjectInfoT `json:"-"`
}

type ObjectInfoT struct {
	Exist       bool
	MD5         string
	Size        int64
	ContentType string
}

func (o *ObjectT) String() string {
	return fmt.Sprintf("{bucket: '%s', object: '%s'}", o.Bucket, o.Path)
}
