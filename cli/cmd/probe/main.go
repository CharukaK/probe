package main

import (
	"github.com/charukak/probe/cli/internal/command/run"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "probe",
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		return
	}
}

func init() {
	RootCmd.AddCommand(run.RunCmd)
}
