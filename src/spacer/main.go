package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var verbose bool

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func Run() {
	self, err := NewSpacer(".")
	if err != nil {
		panic(err)
	}

	verbose = self.GetBool("verbose")
	dockerHost := self.GetString("DOCKER_HOST")
	prefix := self.GetString("prefix")

	platform := NewDockerCompose(dockerHost, prefix)
	deps := self.Dependencies()

	fetchDependencies(self)
	verifyDependencyVersion(self)

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

func verifyDependencyVersion(s *Spacer) (map[string]string, error) {
	var queue []*Spacer
	queue = append(queue, s)

	trace := make(map[string]string)

	for {
		if len(queue) == 0 {
			break
		}

		var x *Spacer
		x, queue = queue[len(queue)-1], queue[:len(queue)-1]

		for _, dep := range x.Dependencies() {
			if _, exist := trace[dep.Name]; exist {
				if trace[dep.Name] == dep.VersionIdentifier() {
					continue
				} else {
					return nil, errors.New("Conflict Dependency Version")
				}
			}

			trace[dep.Name] = dep.VersionIdentifier()
			spacer, err := NewSpacer(dep.Path)
			if err != nil {
				return nil, err
			}
			queue = append(queue, spacer)
		}
	}

	return trace, nil
}

func fetchDependencies(s *Spacer) error {
	var queue []*Spacer
	queue = append(queue, s)

	for {
		if len(queue) == 0 {
			break
		}

		var x *Spacer
		x, queue = queue[len(queue)-1], queue[:len(queue)-1]

		for _, dep := range x.Dependencies() {
			fmt.Println("Initializing", dep.Name)
			fmt.Println("Cloning", dep.LocalPath, dep.RemotePath, "as", dep.Name)
			_, err := dep.Fetch()
			if err != nil && err != ErrLocalPathAlreadyExists {
				return err
			}

			spacer, err := NewSpacer(dep.Path)
			if err != nil {
				return err
			}
			queue = append(queue, spacer)
		}
	}

	return nil
}
