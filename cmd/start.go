package cmd

import (
	"path/filepath"

	spacer "github.com/poga/spacer/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var consumerGroupID string
var env string
var startWithWriteProxy bool
var configFile string

var startCmd = &cobra.Command{
	Use:   "start <projectDirectory>",
	Short: "Start a spacer router for given project",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectDir, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}

		if configFile == "" {
			configFile = filepath.Join(projectDir, "config", "application.yml")
		}

		app, err := spacer.NewApplication(configFile, env)
		if err != nil {
			log.Fatal(err)
		}
		if consumerGroupID != "" {
			app.ConsumerGroupID = consumerGroupID
		}

		err = app.Start(nil, startWithWriteProxy)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	startCmd.Flags().StringVarP(&consumerGroupID, "groupID", "g", "", "Consumer Group ID")
	startCmd.Flags().StringVarP(&env, "env", "e", "", "Environment Name")
	startCmd.Flags().BoolVarP(&startWithWriteProxy, "writeProxy", "w", true, "start with write proxy (set to false if you just want to replay events)")
	startCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	RootCmd.AddCommand(startCmd)
}
