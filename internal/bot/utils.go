package bot

import (
	"os"

	"bot/api/v1alpha1"

	"gopkg.in/yaml.v3"
)

func parseConfig(filepath string) (config v1alpha1.BOTConfigT, err error) {
	configBytes, err := os.ReadFile(filepath)
	if err != nil {
		return config, err
	}

	configBytes = []byte(os.ExpandEnv(string(configBytes)))

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return config, err
	}

	if config.APIService.Address == "" {
		config.APIService.Address = "0.0.0.0"
	}

	return config, err
}
