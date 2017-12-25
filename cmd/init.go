package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init a new spacer project",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO:
		// 1. write nginx configs to target directory
		// TODO: shallow - symlink to nginx configs in the repo
		copyDir()
		// 2. add spacer.yml
		// 3. hello world function

	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}

func copyDir(from string, to string) error {
	from = filepath.Clean(from)
	to = filepath.Clean(to)

	st, err := os.Stat(from)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return errors.Wrap(err, fmt.Sprintf("%s is not a directory", from))
	}

	_, err = os.Stat(to)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		return errors.Wrap(err, fmt.Sprintf("%s already exists", to))
	}

	err = os.MkdirAll(to, st.Mode())
	if err != nil {
		return err
	}

	children, err := ioutil.ReadDir(from)
	if err != nil {
		return err
	}
	for _, c := range children {
		fp := filepath.Join(from, c.Name())
		tp := filepath.Join(to, c.Name())
		if c.IsDir() {
			err = copyDir(fp, tp)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Unable to copy dir %s", fp))
			}
		} else {
			if c.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(fp, tp)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Unable to copy file %s", fp))
			}
		}
	}

	return nil
}

func copyFile(from string, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(to)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	si, err := os.Stat(from)
	if err != nil {
		return err
	}
	err = os.Chmod(to, si.Mode())
	if err != nil {
		return err
	}

	return nil
}
