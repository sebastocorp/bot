package server

import (
	"bot/internal/logger"
	"log"

	"github.com/spf13/cobra"
)

const (
	// FLAG NAMES

	logLevelFlagName = `log-level`

	// ERROR MESSAGES

	logLevelFlagErrMsg = "unable to get flag --log-level: %s"
)

type serverFlagsT struct {
	logLevel string
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

	return err
}
