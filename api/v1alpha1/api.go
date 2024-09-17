package v1alpha1

import "fmt"

type TransferRequestT struct {
	From ObjectT `json:"from"`
	To   ObjectT `json:"to"`
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

type ServerT struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type DatabaseRequestT struct {
	BucketName string `json:"bucket"`
	ObjectPath string `json:"path"`
	MD5        string `json:"md5"`
}

func (o *ObjectT) String() string {
	return fmt.Sprintf("{bucket: '%s', object: '%s'}", o.BucketName, o.ObjectPath)
}
