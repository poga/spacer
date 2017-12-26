package cmd

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
		targetDir, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}
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
		err = spacer.RestoreAssets(targetDir, "app")
		if err != nil {
			log.Fatal(err)
		}
		// 2. add spacer.yml
		err = writeFile(filepath.Join(targetDir, "spacer.yml"), "spacer.example.yml")
		if err != nil {
			log.Fatal(err)
		}

		err = writeFile(filepath.Join(targetDir, ".gitignore"), "appignore")
		if err != nil {
			log.Fatal(err)
		}

		err = writeFile(filepath.Join(targetDir, "nginx.conf"), "nginx.conf")
		if err != nil {
			log.Fatal(err)
		}

		err = os.Mkdir(filepath.Join(targetDir, "logs"), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		err = os.Mkdir(filepath.Join(targetDir, "temp"), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		// 4. git init
		out, err := exec.Command("git", "init", targetDir).Output()
		if err != nil {
			log.Fatal(err)
		}
		log.Info(strings.Trim(string(out), "\n"))
		return
	},
}

func init() {
	initCmd.Flags().StringVarP(&source, "source", "s", "", "Create symlink from source directory instead of copying to target directory. Useful for development")
	RootCmd.AddCommand(initCmd)
}

func writeFile(to string, name string) error {
	data, err := spacer.Asset(name)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(to, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
