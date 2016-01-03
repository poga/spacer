package main

import "github.com/spf13/cobra"

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "start spacer",
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}
