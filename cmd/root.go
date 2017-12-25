package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/poga/spacer/lib"
)

const SpacerVersion = "0.1"

var RootCmd = &cobra.Command{
	Use:   "spacer",
	Short: "serverless platform",
	Long:  `blah`,
}

func run() {
	// TODO: check kafka and openresty exists in path
	app, err := spacer.NewApplication()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Start()
	if err != nil {
		log.Fatal(err)
	}
}
