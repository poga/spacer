package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const SpacerVersion = "0.1"

var rootCmd = &cobra.Command{
	Use:   "spacer",
	Short: "serverless platform",
	Long:  `blah`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start spacer",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "replay specified log",
	Run: func(cmd *cobra.Command, args []string) {
		// use an unique consumer group to replay
		viper.Set("consumer_group_prefix", fmt.Sprintf("spacer-replay-%d", time.Now().Nanosecond()))
		run()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show spacer verision",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(SpacerVersion)
	},
}

func init() {
	rootCmd.AddCommand(startCmd, replayCmd, versionCmd)
}
