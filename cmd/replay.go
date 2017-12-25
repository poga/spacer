package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "replay specified log",
	Run: func(cmd *cobra.Command, args []string) {
		// use an unique consumer group to replay
		viper.Set("consumer_group_prefix", fmt.Sprintf("spacer-replay-%d", time.Now().Nanosecond()))
		run()
	},
}

func init() {
	RootCmd.AddCommand(replayCmd)
}
