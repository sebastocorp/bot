package server

import (
	"bot/internal/logger"
	"context"

	"github.com/spf13/cobra"
)

const (
	descriptionShort = `Execute server process`
	descriptionLong  = `
	Run execute server process`
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "server",
		DisableFlagsInUseLine: true,
		Short:                 descriptionShort,
		Long:                  descriptionLong,

		Run: RunCommand,
	}

	cmd.Flags().String(logLevelFlagName, "info", "Verbosity level for logs")
	cmd.Flags().String(configFlagName, "config.yaml", "Bot service configuration")

	return cmd
}

// RunCommand TODO
// Ref: https://pkg.go.dev/github.com/spf13/pflag#StringSlice
func RunCommand(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	server := serverT{
		context: ctx,
	}
	err := server.getFlags(cmd)
	if err != nil {
		logger.Logger.Fatalf("unable to parse daemon command flags")
	}

	/////////////////////////////
	// EXECUTION FLOW RELATED
	/////////////////////////////

	server.flow()
}
