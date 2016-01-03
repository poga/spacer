package main

import "github.com/spf13/cobra"

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "create spacer configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
