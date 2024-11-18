package v1alpha2

type BOTConfigT struct {
	Name           string                `yaml:"name"`
	LogLevel       string                `yaml:"loglevel"`
	APIService     APIServiceConfigT     `yaml:"apiService"`
	ObjectWorker   ObjectWorkerConfigT   `yaml:"objectWorker"`
	DatabaseWorker DatabaseWorkerConfigT `yaml:"databaseWorker"`
	HashRingWorker HashRingWorkerConfigT `yaml:"hashringWorker,omitempty"`
}

//--------------------------------------------------------------
// API CONFIG
//--------------------------------------------------------------

type APIServiceConfigT struct {
	LogLevel string `yaml:"loglevel"`
	Address  string `yaml:"address"`
	Port     string `yaml:"port"`
}

//--------------------------------------------------------------
// OBJECT STORAGE WORKER CONFIG
//--------------------------------------------------------------

type ObjectWorkerConfigT struct {
	LogLevel              string                    `yaml:"loglevel"`
	MaxChildTheads        int                       `yaml:"maxChildTheads,omitempty"`
	RequestsByChildThread int                       `yaml:"requestsByChildThread,omitempty"`
	ObjectStorage         ObjectStorageT            `yaml:"objectStorage"`
	ObjectModification    ObjectModificationConfigT `yaml:"objectModification"`
}

type ObjectStorageT struct {
	S3  S3T  `yaml:"s3"`
	GCS GCST `yaml:"gcs"`
}

type S3T struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	Region          string `yaml:"region,omitempty"`
	Secure          bool   `yaml:"secure,omitempty"`
}

type GCST struct {
	CredentialsFile string `yaml:"credentialsFile"`
}

type ObjectModificationConfigT struct {
	Type string                         `yaml:"type"`
	Mods map[string]ModificationConfigT `yaml:"modifications"`
}

type ModificationConfigT struct {
	Bucket       string `yaml:"bucket"`
	AddPrefix    string `yaml:"addPrefix"`
	RemovePrefix string `yaml:"removePrefix"`
}

//--------------------------------------------------------------
// DATABASE WORKER CONFIG
//--------------------------------------------------------------

type DatabaseWorkerConfigT struct {
	LogLevel              string    `yaml:"loglevel"`
	MaxChildTheads        int       `yaml:"maxChildTheads,omitempty"`
	RequestsByChildThread int       `yaml:"requestsByChildThread,omitempty"`
	Database              DatabaseT `yaml:"database"`
}

type DatabaseT struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Table    string `yaml:"table"`
}

//--------------------------------------------------------------
// HASHRING WORKER CONFIG
//--------------------------------------------------------------

type HashRingWorkerConfigT struct {
	Enabled  bool   `yaml:"enabled"`
	LogLevel string `yaml:"loglevel"`
	Proxy    string `yaml:"proxy"`
	VNodes   int    `yaml:"vnodes"`
}
