package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "spacer",
	Short: "the microservice dependcency manager",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UsageString())
	},
}

func init() {
	RootCmd.AddCommand(InitCmd)
	RootCmd.AddCommand(RunCmd)
}
