package v1alpha1

type BOTConfigT struct {
	Name           string                `yaml:"name"`
	APIService     APIServiceConfigT     `yaml:"apiService"`
	ObjectWorker   ObjectWorkerConfigT   `yaml:"objectStorageWorker"`
	DatabaseWorker DatabaseWorkerConfigT `yaml:"databaseWorker"`
	HashRingWorker HashRingWorkerConfigT `yaml:"hashringWorker,omniempty"`
}

//--------------------------------------------------------------
// API CONFIG
//--------------------------------------------------------------

type APIServiceConfigT struct {
	Port    string `yaml:"port"`
	Address string `yaml:"address"`
}

//--------------------------------------------------------------
// OBJECT STORAGE WORKER CONFIG
//--------------------------------------------------------------

type ObjectWorkerConfigT struct {
	ParallelRequests int            `yaml:"parallelRequests"`
	ObjectStorage    ObjectStorageT `yaml:"objectStorage"`
}

type ObjectStorageT struct {
	S3  S3T  `yaml:"s3"`
	GCS GCST `yaml:"gcs"`
}

type S3T struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyId"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	Region          string `yaml:"region,omniempty"`
	Secure          bool   `yaml:"secure,omniempty"`
}

type GCST struct {
	CredentialsFile string `yaml:"credentialsFile"`
}

//--------------------------------------------------------------
// DATABASE WORKER CONFIG
//--------------------------------------------------------------

type DatabaseWorkerConfigT struct {
	ParallelRequests    int       `yaml:"parallelRequests,omniempty"`
	InsertsByConnection int       `yaml:"insertsByConnection,omniempty"`
	Database            DatabaseT `yaml:"database"`
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
	Enabled bool   `yaml:"enabled"`
	Proxy   string `yaml:"proxy"`
	VNodes  int    `yaml:"vnodes"`
}
