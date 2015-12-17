package main

import (
	"os"
	"os/exec"
)

// Exec will execute give command and respect verbose config
func Exec(name string, c ...string) ([]byte, error) {
	cmd := exec.Command(name, c...)
	if verbose {
		cmd.Env = os.Environ()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		return nil, err
	}

	return cmd.CombinedOutput()
}
