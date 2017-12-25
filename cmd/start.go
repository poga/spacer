package cmd

import "github.com/spf13/cobra"

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start spacer",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: also start nginx
		// TODO: tail error.log and access.log
		// TODO: parse error.log and access.log?
		run()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
