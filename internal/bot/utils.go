package bot

import (
	"fmt"
	"os"

	"bot/api/v1alpha2"

	"gopkg.in/yaml.v3"
)

func parseConfig(filepath string) (config v1alpha2.BOTConfigT, err error) {
	configBytes, err := os.ReadFile(filepath)
	if err != nil {
		return config, err
	}

	configBytes = []byte(os.ExpandEnv(string(configBytes)))

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return config, err
	}

	return config, err
}

func (b *BotT) checkConfig() (err error) {

	//--------------------------------------------------------------
	// CHECK API CONFIG
	//--------------------------------------------------------------

	if b.config.APIService.Address == "" {
		b.config.APIService.Address = "0.0.0.0"
	}

	//--------------------------------------------------------------
	// CHECK OBJECT CONFIG
	//--------------------------------------------------------------

	if b.config.ObjectWorker.MaxChildTheads <= 0 {
		err = fmt.Errorf("config option objectWorker.maxChildTheads must be a number > 0")
		return err
	}

	if b.config.ObjectWorker.RequestsByChildThread <= 0 {
		err = fmt.Errorf("config option objectWorker.requestsByChildThread must be a number > 0")
		return err
	}

	//--------------------------------------------------------------
	// CHECK DATABASE CONFIG
	//--------------------------------------------------------------

	if b.config.DatabaseWorker.Database.Host == "" {
		err = fmt.Errorf("database host config is empty")
		return err
	}

	if b.config.DatabaseWorker.Database.Port == "" {
		err = fmt.Errorf("database port config is empty")
		return err
	}

	if b.config.DatabaseWorker.Database.Database == "" {
		err = fmt.Errorf("database name config is empty")
		return err
	}

	if b.config.DatabaseWorker.Database.Table == "" {
		err = fmt.Errorf("database table config is empty")
		return err
	}

	if b.config.DatabaseWorker.Database.Username == "" {
		err = fmt.Errorf("database user config is empty")
		return err
	}

	if b.config.DatabaseWorker.Database.Password == "" {
		err = fmt.Errorf("database password config is empty")
		return err
	}

	if b.config.DatabaseWorker.MaxChildTheads <= 0 {
		err = fmt.Errorf("config option databaseWorker.maxChildTheads must be a number > 0")
		return err
	}

	if b.config.DatabaseWorker.RequestsByChildThread <= 0 {
		err = fmt.Errorf("config option databaseWorker.requestsByChildThread must be a number > 0")
		return err
	}

	//--------------------------------------------------------------
	// CHECK HASHRING CONFIG
	//--------------------------------------------------------------

	if b.config.HashRingWorker.Enabled {
		if b.config.HashRingWorker.VNodes <= 0 {
			b.config.HashRingWorker.VNodes = 1
		}
	}

	return err
}
