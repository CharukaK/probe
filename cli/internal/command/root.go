package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Short: "CLI client for handling probe files",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run")
	},
}

