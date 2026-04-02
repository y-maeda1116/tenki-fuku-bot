package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/y-maeda1116/template-go-cross/internal/core"
)

func NewHelloCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "hello",
		Short: "Say hello",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" {
				name = "World"
			}
			svc := core.NewService()
			msg, err := svc.SayHello(name)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err) //nolint:errcheck
				return
			}
			fmt.Println(msg)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name to greet")

	return cmd
}
