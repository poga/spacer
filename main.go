package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var verbose bool

func main() {
	self, err := NewSpacer(".")
	if err != nil {
		panic(err)
	}

	verbose = self.GetBool("verbose")
	dockerHost := self.GetString("DOCKER_HOST")
	prefix := self.GetString("prefix")

	platform := NewDockerCompose(dockerHost, prefix)
	deps := self.Dependencies()

	for _, dep := range deps {
		fmt.Println("Initializing", dep.Name)
		fmt.Println("Cloning", dep.LocalPath, dep.RemotePath, "as", dep.Name)
		_, err := dep.Fetch()
		if err != nil {
			if err != ErrLocalPathAlreadyExists {
				log.Panic(err)
			}
		}
	}

	for _, dep := range deps {
		// docker-compose build && docker-compose up
		fmt.Println("Building", platform.ConfigPath(dep), "...")
		out, err := platform.Build(dep)
		if err != nil {
			fmt.Println(string(out))
			log.Panic(err)
		}
	}

	for _, dep := range deps {
		fmt.Println("Starting", platform.ConfigPath(dep), "...")
		platform.Start(dep)
	}

	// setup a proxy for each service
	for serviceName, exposeURL := range platform.Running() {
		prefix, proxy := NewProxy(serviceName, exposeURL)
		http.HandleFunc(prefix, proxy.ServeHTTP)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			fmt.Println("Stopping services...")
			for _, dep := range deps {
				fmt.Println("Stopping", dep.Name, "...")
				platform.Stop(dep)
			}
			os.Exit(0)
		}
	}()

	fmt.Println("Spacer is ready and rocking at " + self.GetString("listen"))
	http.ListenAndServe(self.GetString("listen"), nil)
}
