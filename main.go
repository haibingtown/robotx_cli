package main

import (
	"os"

	"github.com/haibingtown/robotx_cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(cmd.HandleError(err))
	}
}
