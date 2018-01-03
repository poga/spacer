package cmd

import (
	"path/filepath"

	spacer "github.com/poga/spacer/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var configName string
var consumerGroupID string

var startCmd = &cobra.Command{
	Use:   "start <projectDirectory> <env>",
	Short: "Start a spacer router for given project",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectDir, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}
		env := args[1]

		app, err := spacer.NewApplication(projectDir, configName, env)
		if err != nil {
			log.Fatal(err)
		}
		if consumerGroupID != "" {
			app.ConsumerGroupID = consumerGroupID
		}

		err = app.Start()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	startCmd.Flags().StringVarP(&configName, "config", "c", "spacer", "Config Filename")
	startCmd.Flags().StringVarP(&consumerGroupID, "groupID", "g", "", "Consumer Group ID")
	RootCmd.AddCommand(startCmd)
}
