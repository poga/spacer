package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type Service struct {
	Name string
	Root string
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
	return exec.Command("docker-compose", "-f", s.ConfigPath(), "up").Start()
}

func (s Service) Stop() ([]byte, error) {
	fmt.Println("Stopping", s.ConfigPath())
	return exec.Command("docker-compose", "-f", s.ConfigPath(), "stop").CombinedOutput()
}

func (s Service) GetExposedURL(serviceName string, port string) (string, error) {
	output, err := exec.Command("docker-compose", "-f", s.ConfigPath(), "port", serviceName, port).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\n"), nil
}
