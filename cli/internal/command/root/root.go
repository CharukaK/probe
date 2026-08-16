package command

import (
	"github.com/charukak/probe/cli/internal/command/run"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "probe",

}

func init() {
	RootCmd.AddCommand(run.RunCmd)
}
