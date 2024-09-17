package global

import (
	"os"

	"bot/api/v1alpha1"
	"bot/internal/managers/hashring"
	"bot/internal/pools"

	"gopkg.in/yaml.v3"
)

const (
	HeaderContentTypeAppJson = "application/json"

	EndpointHealth          = "/health"
	EndpointInfo            = "/info"
	EndpointRequestObject   = "/transfer"
	EndpointRequestDatabase = "/request/database"
)

var Config v1alpha1.BOTConfigT

var HashRing *hashring.HashRingT

var ServerInstancesPool = pools.NewServerPool()

var TransferRequestPool = pools.NewTransferRequestPool()

var DatabaseRequestPool = pools.NewDatabaseRequestPool()

func ParseConfig(filepath string) (err error) {
	configBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	configBytes = []byte(os.ExpandEnv(string(configBytes)))

	err = yaml.Unmarshal(configBytes, &Config)
	if err != nil {
		return err
	}

	return err
}
