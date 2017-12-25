package cmd

import (
	"github.com/spf13/cobra"
)

const SpacerVersion = "0.1"

var RootCmd = &cobra.Command{
	Use:   "spacer",
	Short: "serverless platform",
	Long:  `blah`,
}
