package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"

	spacer "github.com/poga/spacer/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var source string

var initCmd = &cobra.Command{
	Use:   "init [targetDir]",
	Short: "init a new spacer project",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetDir := args[0]
		// TODO:
		// 1. write nginx configs to target directory
		if targetDir == "" {
			log.Fatalf("Target Directory is Required")
		}
		if source != "" {
			err := os.Symlink(source, targetDir)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		err := spacer.RestoreAssets(targetDir, "nginx")
		if err != nil {
			log.Fatal(err)
		}
		// 2. add spacer.yml
		spacerYAML, err := spacer.Asset("spacer.example.yml")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(targetDir, "spacer.yml"), spacerYAML, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return
		// 3. hello world function
	},
}

func init() {
	initCmd.Flags().StringVarP(&source, "source", "s", "", "Create symlink from source directory instead of copying to target directory. Useful for development")
	RootCmd.AddCommand(initCmd)
}
