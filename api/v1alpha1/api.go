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

func (t *TransferRequestT) String() string {
	return fmt.Sprintf("{from: {bucket: '%s', object: '%s'} to: {bucket: '%s', object: '%s'}}", t.From.BucketName, t.From.ObjectPath, t.To.BucketName, t.To.ObjectPath)
}

func (d *DatabaseRequestT) String() string {
	return fmt.Sprintf("{bucket: '%s', object: '%s'}", d.BucketName, d.ObjectPath)
}

func (s *ServerT) String() string {
	return fmt.Sprintf("{name: '%s', adress: '%s'}", s.Name, s.Address)
}

func (o *ObjectT) String() string {
	return fmt.Sprintf("{bucket: '%s', object: '%s'}", o.BucketName, o.ObjectPath)
}
