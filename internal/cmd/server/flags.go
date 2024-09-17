package server

import (
	"bot/internal/logger"
	"log"

	"github.com/spf13/cobra"
)

const (
	// FLAG NAMES

	logLevelFlagName = `log-level`
	configFlagName   = `config`

	// ERROR MESSAGES

	logLevelFlagErrMsg = "unable to get flag --log-level: %s"
	configFlagErrMsg   = "unable to get flag --config: %s"
)

type serverFlagsT struct {
	logLevel string
	config   string
}

func (d *serverT) getFlags(cmd *cobra.Command) (err error) {

	// Get root command flags
	d.flags.logLevel, err = cmd.Flags().GetString(logLevelFlagName)
	if err != nil {
		log.Fatalf(logLevelFlagErrMsg, err.Error())
	}

	level, err := logger.GetLevel(d.flags.logLevel)
	if err != nil {
		log.Fatalf(logLevelFlagErrMsg, err.Error())
	}

	logger.InitLogger(d.context, level)

	// Get server command flags

	d.flags.config, err = cmd.Flags().GetString(configFlagName)
	if err != nil {
		log.Fatalf(configFlagErrMsg, err.Error())
	}

	return err
}
