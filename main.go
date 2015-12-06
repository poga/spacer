package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/spf13/viper"
)

func main() {
	dockerHost := viper.GetString("DOCKER_HOST")
	prefix := viper.GetString("prefix")
	platform := NewDockerCompose(dockerHost, prefix)
	deps := getDependencies()

	for _, dep := range deps {
		fmt.Println("Initializing", dep.Name)
		fmt.Println("\tCloning", dep.LocalPath, "...")
		out, err := dep.Fetch()
		if err != nil {
			if err != ErrLocalPathAlreadyExists {
				log.Panic(err)
			}
		}

		// docker-compose build && docker-compose up
		fmt.Println("\tBuilding", platform.ConfigPath(dep), "...")
		out, err = platform.Build(dep)
		if err != nil {
			fmt.Println(string(out))
			log.Panic(err)
		}

		fmt.Println("\tStarting", platform.ConfigPath(dep), "...")
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
				fmt.Println("\tStopping", dep.Name, "...")
				platform.Stop(dep)
			}
			os.Exit(0)
		}
	}()

	fmt.Println("Spacer is ready and rocking at " + viper.GetString("listen"))
	http.ListenAndServe(viper.GetString("listen"), nil)
}
