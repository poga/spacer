package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

func main() {
	dockerHost := os.Getenv("DOCKER_HOST")
	var services []Service

	spacerfile, err := ioutil.ReadFile("Spacerfile")
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(spacerfile), "\n")
	for _, l := range lines {
		if l == "" {
			continue
		}
		s, err := NewService("services", l, dockerHost)
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
		fmt.Println("\tBuilding", s.ConfigPath(), "...")
		_, err = s.Build()
		if err != nil {
			log.Panic(err)
		}

		fmt.Println("\tStarting", s.ConfigPath(), "...")
		s.Start()
	}

	// setup a proxy for each service
	for _, s := range services {
		// TODO not just web
		prefix, proxy := NewProxy(s, "web")
		http.HandleFunc(prefix, proxy.ServeHTTP)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			fmt.Println("Stopping services...")
			for _, s := range services {
				fmt.Println("\tStopping", s.Name, "...")
				output, err := s.Stop()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(output))
			}
			os.Exit(0)
		}
	}()

	fmt.Println("Spacer is ready and rocking at 0.0.0.0:9064")
	http.ListenAndServe(":9064", nil)
}
