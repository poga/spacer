package main

import (
	"os"
	"os/exec"

	"github.com/spf13/viper"
)

func Exec(name string, c ...string) ([]byte, error) {
	verbose := viper.GetBool("verbose")

	cmd := exec.Command(name, c...)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		return nil, err
	}

	return cmd.CombinedOutput()
}
