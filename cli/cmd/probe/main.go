package main

import (
	"fmt"

	command "github.com/charukak/probe/cli/internal/command/root"
)

func main() {
	if err := command.RootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
