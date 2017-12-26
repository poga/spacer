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
	Use:   "start [flags] [projectDirectory]",
	Short: "Start a spacer router for given project",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectDir, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}

		app, err := spacer.NewApplication(projectDir, configName)
		if err != nil {
			log.Fatal(err)
		}
		if consumerGroupID != "" {
			app.Set("consumerGroupPrefix", consumerGroupID)
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
