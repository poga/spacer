package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Service struct {
	Name        string
	Root        string
	ExposedURLs map[string]*url.URL
}

func NewService(path string) Service {
	s := Service{
		Name:        filepath.Base(path),
		Root:        path,
		ExposedURLs: make(map[string]*url.URL),
	}

	return s
}

func (s Service) ConfigPath() string {
	return s.Root + "/docker-compose.yml"
}

func (s Service) Build() ([]byte, error) {
	fmt.Println("Building", s.ConfigPath())
	return exec.Command("docker-compose", "-f", s.ConfigPath(), "build").CombinedOutput()
}

func (s Service) Start() error {
	fmt.Println("Starting", s.ConfigPath())
	err := exec.Command("docker-compose", "-f", s.ConfigPath(), "up").Start()
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	// loop through docker-compose.yml and look for exposed ip, save it
	ymlBytes, err := ioutil.ReadFile(s.ConfigPath())
	if err != nil {
		log.Panic(err)
	}
	m := make(map[string]interface{})
	err = yaml.Unmarshal(ymlBytes, &m)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("%v\n", m)

	for serviceName, conf := range m {
		if ports, ok := conf.(map[interface{}]interface{})["ports"]; ok {
			// TODO handle multiple exposed ports
			portValue := ports.([]interface{})[0].(string)
			var innerPort string

			// TODO handle 127.0.0.1:8001:8001 style config
			if len(strings.Split(portValue, ":")) == 2 {
				innerPort = strings.Split(portValue, ":")[1]
			} else {
				innerPort = portValue
			}
			output, err := s.getExposedURL(serviceName, innerPort)
			if err != nil {
				log.Panic(err)
			}
			u, err := url.Parse("http://" + output)
			if err != nil {
				log.Panic(err)
			}
			s.ExposedURLs[serviceName] = u
		}
	}

	exec.Command("docker-compose", "-f", s.ConfigPath(), "up").Start()
	return nil
}

func (s Service) Stop() ([]byte, error) {
	fmt.Println("Stopping", s.ConfigPath())
	return exec.Command("docker-compose", "-f", s.ConfigPath(), "stop").CombinedOutput()
}

func (s Service) getExposedURL(serviceName string, port string) (string, error) {
	fmt.Println("getExposedURL", serviceName, port, s.ConfigPath())
	output, err := exec.Command("docker-compose", "-f", s.ConfigPath(), "port", serviceName, port).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\n"), nil
}
