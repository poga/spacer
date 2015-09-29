package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Service struct {
	Name        string
	Path        string
	ExposedURLs map[string]*url.URL
	Repo        repo
	Prefix      string
	RunInfo     runInfo
}

type repo struct {
	username string
	reponame string
}

type runInfo struct {
	Host string
}

func (s Service) LocalRepoPath() string {
	return strings.Join([]string{s.Prefix, s.Repo.username, s.Repo.reponame}, "/")
}

func NewService(prefix string, path string, dockerHost string) (Service, error) {
	os.Mkdir(prefix, 0777)

	var host string
	if dockerHost != "" {
		var err error
		dockerHostURL, err := url.Parse(dockerHost)
		if err != nil {
			log.Panic(err)
		}
		host, _, err = net.SplitHostPort(dockerHostURL.Host)
		if err != nil {
			log.Panic(err)
		}
	} else {
		host = ""
	}

	if len(strings.Split(path, "/")) < 2 {
		return Service{}, errors.New("Unknown repo format: " + path)
	}
	s := Service{
		Name:        filepath.Base(path),
		Path:        strings.Join([]string{prefix, path}, "/"),
		ExposedURLs: make(map[string]*url.URL),
		Prefix:      prefix,
		RunInfo: runInfo{
			Host: host,
		},
	}
	username := strings.Split(path, "/")[0]
	reponame := strings.Split(path, "/")[1]
	s.Repo = repo{username, reponame}

	return s, nil
}

var ErrLocalPathAlreadyExists = errors.New("local path already exists")

func (s Service) Clone() error {
	output, err := exec.Command("git", "clone", s.RepoCloneURL(), s.LocalRepoPath()).CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "already exists and is not an empty directory") {
			return ErrLocalPathAlreadyExists
		} else {
			return err
		}
	}

	return nil
}

func (s Service) RepoCloneURL() string {
	return "git@github.com:" + s.Repo.username + "/" + s.Repo.reponame + ".git"
}

func (s Service) ConfigPath() string {
	return s.Path + "/docker-compose.yml"
}

func (s Service) Build() ([]byte, error) {
	return exec.Command("docker-compose", "-f", s.ConfigPath(), "build").CombinedOutput()
}

func (s Service) Start() error {
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
			urlStr, err := s.getExposedURLString(serviceName, innerPort)
			// fmt.Println(output)
			if err != nil {
				log.Panic(err)
			}
			u, err := url.Parse("http://" + urlStr)
			if err != nil {
				log.Panic(err)
			}
			if s.RunInfo.Host != "" {
				_, originPort, err := net.SplitHostPort(u.Host)
				if err != nil {
					return err
				}
				u.Host = s.RunInfo.Host + ":" + originPort
			}
			fmt.Println(u.Host)
			s.ExposedURLs[serviceName] = u
		}
	}

	return nil
}

func (s Service) Stop() ([]byte, error) {
	return exec.Command("docker-compose", "-f", s.ConfigPath(), "stop").CombinedOutput()
}

func (s Service) getExposedURLString(serviceName string, port string) (string, error) {
	output, err := exec.Command("docker-compose", "-f", s.ConfigPath(), "port", serviceName, port).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\n"), nil
}
