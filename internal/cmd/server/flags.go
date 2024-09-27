package server

import (
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
	config string
}

func getFlags(cmd *cobra.Command) (flags serverFlagsT, err error) {

	// Get server command flags

	flags.config, err = cmd.Flags().GetString(configFlagName)
	if err != nil {
		log.Fatalf(configFlagErrMsg, err.Error())
	}

	return flags, err
}
