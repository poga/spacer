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
	InitEtcd()

	dockerHost := viper.GetString("DOCKER_HOST")
	platform := NewDockerCompose(dockerHost)
	var services []Service

	for _, dep := range GetDeps() {
		l := dep.Repo
		if l == "" {
			continue
		}
		s, err := NewService("services", l)
		if err != nil {
			log.Panic(err)
		}
		fmt.Println("Initializing", s.Name)
		fmt.Println("\tCloning", s.RepoCloneURL(), "into", s.LocalRepoPath(), "...")
		err = s.Clone()
		if err != nil {
			if err != ErrLocalPathAlreadyExists {
				log.Panic(err)
			} else {
				fmt.Println("\tService already exists: " + s.LocalRepoPath())
			}
		}
		services = append(services, s)

		// docker-compose build && docker-compose up
		fmt.Println("\tBuilding", platform.ConfigPath(s), "...")
		err = platform.Build(s)
		if err != nil {
			log.Panic(err)
		}

		fmt.Println("\tStarting", platform.ConfigPath(s), "...")
		platform.Start(s)
	}

	// setup a proxy for each service
	for _, s := range services {
		// TODO not just web
		prefix, proxy := NewProxy(s)
		http.HandleFunc(prefix, proxy.ServeHTTP)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			fmt.Println("Stopping services...")
			for _, s := range services {
				fmt.Println("\tStopping", s.Name, "...")
				platform.Stop(s)
			}
			os.Exit(0)
		}
	}()

	fmt.Println("Spacer is ready and rocking at 0.0.0.0:9064")
	http.ListenAndServe(":9064", nil)
}
