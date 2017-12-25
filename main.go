package main

import (
	"log"

	"github.com/poga/spacer/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
