package main

import (
	"log/slog"
	"os"

	"github.com/charukak/probe/cli/internal/command/run"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "probe",
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		slog.Error("error executing command", "error", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(run.RunCmd)
}
