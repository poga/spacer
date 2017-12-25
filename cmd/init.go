package cmd

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init a new spacer project",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO:
		// 1. copy nginx configs to target directory
		// 2. add spacer.yml
		// 3. hello world function
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
