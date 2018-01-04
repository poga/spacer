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

		if _, err := os.Stat(targetDir); err == nil {
			log.Fatal("Target Directory already exists")
		}

		err = series([]func() error{
			func() error { return spacer.RestoreAssets(targetDir, "app") },
			func() error { return spacer.RestoreAssets(targetDir, "bin") },
			func() error { return spacer.RestoreAssets(targetDir, "config") },
			func() error { return writeFile(filepath.Join(targetDir, ".gitignore"), "appignore") },
			func() error { return os.Mkdir(filepath.Join(targetDir, "logs"), os.ModePerm) },
			func() error { return os.Mkdir(filepath.Join(targetDir, "temp"), os.ModePerm) },
			func() error { return os.Mkdir(filepath.Join(targetDir, "test"), os.ModePerm) },
			func() error { return writeFile(filepath.Join(targetDir, "test/hello.t.md"), "hello.t.md") },
		})
		if err != nil {
			log.Fatal(err)
		}

		out, err := exec.Command("git", "init", targetDir).CombinedOutput()
		if err != nil {
			log.Infof(string(out))
			log.Fatal(err)
		}
		return
	},
}

func init() {
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
