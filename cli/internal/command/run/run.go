package run

import (
	"fmt"
	"os"

	"github.com/charukak/probe/cli/internal/parser"
	"github.com/charukak/probe/cli/internal/request"
	"github.com/spf13/cobra"
)


var RunCmd = &cobra.Command{
	Use:   "run <file-path>",
	Short: "Run .probe file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		source, err := os.ReadFile(args[0])

		if err != nil {
			return err
		}

		sourceFile, err := parser.Parse(source)
		if err != nil {
			return err
		}

		fmt.Println(fmt.Sprintf("%v", sourceFile))
		err = request.BuildAndExec(sourceFile, &request.Opts{Verbose: verbose})

		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	RunCmd.Flags().BoolP("verbose", "v", false, "log HTTP request/response detail to stdout")
}
