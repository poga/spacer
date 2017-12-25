package cmd

import (
	spacer "github.com/poga/spacer/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start spacer",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: also start nginx
		// TODO: tail error.log and access.log
		// TODO: parse error.log and access.log?

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
	RootCmd.AddCommand(startCmd)
}
