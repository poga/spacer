package cmd

import (
	"io/ioutil"
	"os"
	"os/exec"
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
		// Initialize a spacer project from defined template and setup directory structure for nginx
		targetDir, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}

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

		err = series([]func() error{
			func() error { return writeFile(filepath.Join(targetDir, "spacer.yml"), "spacer.example.yml") },
			func() error { return writeFile(filepath.Join(targetDir, ".gitignore"), "appignore") },
			func() error { return writeFile(filepath.Join(targetDir, "nginx.conf"), "nginx.conf") },
			func() error { return os.Mkdir(filepath.Join(targetDir, "logs"), os.ModePerm) },
			func() error { return os.Mkdir(filepath.Join(targetDir, "temp"), os.ModePerm) },
		})
		if err != nil {
			log.Fatal(err)
		}

		_, err = exec.Command("git", "init", targetDir).Output()
		if err != nil {
			log.Fatal(err)
		}
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

func series(funcs []func() error) error {
	for _, f := range funcs {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}
