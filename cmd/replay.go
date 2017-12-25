package cmd

import (
	"fmt"
	"time"

	spacer "github.com/poga/spacer/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "replay specified log",
	Run: func(cmd *cobra.Command, args []string) {
		// use an unique consumer group to replay
		viper.Set("consumer_group_prefix", fmt.Sprintf("spacer-replay-%d", time.Now().Nanosecond()))

		// TODO: check kafka and openresty exists in path
		app, err := spacer.NewApplication()
		if err != nil {
			log.Fatal(err)
		}

		err = app.Start()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(replayCmd)
}
