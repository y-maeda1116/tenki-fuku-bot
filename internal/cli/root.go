package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "myapp",
		Short: "A CLI and Desktop application",
		Long:  "MyApp is a CLI and Desktop application template built with Go and Wails.",
	}

	cmd.AddCommand(NewVersionCommand(version))
	cmd.AddCommand(NewHelloCommand())

	return cmd
}
