package cmd

import (
	"bot/internal/cmd/server"
	"bot/internal/cmd/version"

	"github.com/spf13/cobra"
)

const (
	descriptionShort = `TODO`
	descriptionLong  = `
	TODO`
)

func NewRootCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: descriptionShort,
		Long:  descriptionLong,
	}

	cmd.AddCommand(
		version.NewCommand(),
		server.NewCommand(),
	)

	return cmd
}
