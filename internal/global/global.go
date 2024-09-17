package global

import (
	"bot/api/v1alpha1"
	"bot/internal/hashring"
	"bot/internal/pools"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	HeaderContentTypeAppJson = "application/json"
)

var ServerConfig v1alpha1.BOTConfigT

var ServerReference pools.ServerT

var HashRing *hashring.HashRingT

var ServerInstancesPool = pools.NewServerInstancesPool()

var TransferRequestPool = pools.NewTransferRequestPool()

var DatabaseRequestPool = pools.NewDatabaseRequestPool()

func ParseConfig(filepath string) (err error) {
	configBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	configBytes = []byte(os.ExpandEnv(string(configBytes)))

	err = yaml.Unmarshal(configBytes, &ServerConfig)
	if err != nil {
		return err
	}

	return err
}
