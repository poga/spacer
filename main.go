package main

import (
	"os"
	"strings"

	"github.com/poga/spacer/cmd"

	log "github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
	}
}
