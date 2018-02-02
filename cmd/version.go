package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show spacer verision",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(SpacerVersion)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
