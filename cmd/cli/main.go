package main

import (
	"os"

	"github.com/y-maeda1116/template-go-cross/internal/cli"
	"github.com/y-maeda1116/template-go-cross/internal/version"
)

func main() {
	cmd := cli.NewRootCommand(version.Version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
