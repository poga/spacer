// A data-driven test runner
package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	spacer "github.com/poga/spacer/pkg"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [projectDirectory]",
	Short: "Run test specs",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectDir, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}

		app, err := spacer.NewApplication(projectDir, configName, "test")
		if err != nil {
			log.Fatal(err)
		}
		app.ConsumerGroupID = fmt.Sprintf("spacer-test-%s-%d", app.Name(), time.Now().Unix())

		appReady := make(chan int)

		go func() {
			err = app.Start(appReady)
			if err != nil {
				log.Fatal(err)
			}
		}()

		<-appReady
		// TODO: parse test spec
		spacer.RunTest(app.Delegator(), filepath.Join(projectDir, "test"))

		// TODO: send http request
		// TODO: assert response
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
}
