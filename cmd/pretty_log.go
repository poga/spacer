package cmd

import (
	"path/filepath"

	"github.com/hpcloud/tail"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var prettyLogCmd = &cobra.Command{
	Use:   "pretty-log [projectDirectory]",
	Short: "Show access.log and error.log",
	Run: func(cmd *cobra.Command, args []string) {
		projectDir, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}

		// access.log
		go func() {
			t, err := tail.TailFile(filepath.Join(projectDir, "logs", "access.log"), tail.Config{Follow: true})
			if err != nil {
				log.Fatal(err)
			}
			for line := range t.Lines {
				log.Info(line.Text)
			}
		}()

		// error.log
		t, err := tail.TailFile(filepath.Join(projectDir, "logs", "error.log"), tail.Config{Follow: true})
		if err != nil {
			log.Fatal(err)
		}
		for line := range t.Lines {
			log.Error(line.Text)
		}
	},
}

func init() {
	RootCmd.AddCommand(prettyLogCmd)
}
