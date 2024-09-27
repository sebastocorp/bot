package server

import (
	"bot/internal/bot"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	flags, err := getFlags(cmd)
	if err != nil {
		log.Fatalf("unable to parse daemon command flags")
	}

	/////////////////////////////
	// EXECUTION FLOW RELATED
	/////////////////////////////
	botServer, err := bot.NewBotServer(flags.config)
	if err != nil {
		log.Fatalf("unable to config bot server: %s", err.Error())
	}

	// create channels to manage shutdown actions
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	shutdownActionsDone := make(chan bool, 1)
	go botServer.ShutdownActions(shutdownActionsDone, signals)

	botServer.Run()

	<-shutdownActionsDone
}
