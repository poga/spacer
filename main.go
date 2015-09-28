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
	var services []Service

	spacerfile, err := ioutil.ReadFile("Spacerfile")
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(spacerfile), "\n")
	for _, l := range lines {
		s, err := NewService("services", l)
		if err != nil {
			continue
		}
		err = s.Clone()
		if err != nil {
			if err != ErrLocalPathAlreadyExists {
				log.Panic(err)
			} else {
				fmt.Println("Service already exists: " + s.LocalRepoPath())
			}
		}
		services = append(services, s)

		// docker-compose build && docker-compose up
		dcb, err := s.Build()
		fmt.Println(string(dcb))
		if err != nil {
			log.Panic(err)
		}

		s.Start()
	}

	// setup a proxy for each service
	for _, s := range services {
		// TODO not just web
		prefix, proxy := s.ReverseProxy("web")
		http.HandleFunc(prefix, proxy.ServeHTTP)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			fmt.Println("\nReceived an interrupt, stopping services...")
			for _, s := range services {
				output, err := s.Stop()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(output))
			}
			os.Exit(0)
		}
	}()

	fmt.Println("\nSpacer is ready and rocking at 0.0.0.0:9064")
	http.ListenAndServe(":9064", nil)
}
