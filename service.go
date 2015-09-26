package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
}

type repo struct {
	username string
	reponame string
}

func (s Service) LocalRepoPath() string {
	return strings.Join([]string{s.Prefix, s.Repo.username, s.Repo.reponame}, "/")
}

func NewService(prefix string, path string) (Service, error) {
	os.Mkdir(prefix, 0777)

	if len(strings.Split(path, "/")) < 2 {
		return Service{}, errors.New("Unknown repo format: " + path)
	}
	s := Service{
		Name:        filepath.Base(path),
		Path:        strings.Join([]string{prefix, path}, "/"),
		ExposedURLs: make(map[string]*url.URL),
		Prefix:      prefix,
	}
	username := strings.Split(path, "/")[0]
	reponame := strings.Split(path, "/")[1]
	s.Repo = repo{username, reponame}

	return s, nil
}

var ErrLocalPathAlreadyExists = errors.New("local path already exists")

func (s Service) ReverseProxy(composeName string) (string, *httputil.ReverseProxy) {
	prefix := "/" + s.Name
	proxy := httputil.NewSingleHostReverseProxy(s.ExposedURLs[composeName])
	proxy.Director = direct(prefix, s.ExposedURLs[composeName])
	return prefix, proxy
}

func (s Service) Clone() error {
	url := "git@github.com:" + s.Repo.username + "/" + s.Repo.reponame + ".git"
	fmt.Println("Cloning", url, "into", s.LocalRepoPath(), "...")
	output, err := exec.Command("git", "clone", url, s.LocalRepoPath()).CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "already exists and is not an empty directory") {
			return ErrLocalPathAlreadyExists
		} else {
			return err
		}
	}

	return nil
}

func (s Service) ConfigPath() string {
	return s.Path + "/docker-compose.yml"
}

func (s Service) Build() ([]byte, error) {
	fmt.Println("Building", s.ConfigPath(), "...")
	return exec.Command("docker-compose", "-f", s.ConfigPath(), "build").CombinedOutput()
}

func (s Service) Start() error {
	fmt.Println("Starting", s.ConfigPath(), "...")
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
			output, err := s.getExposedURL(serviceName, innerPort)
			fmt.Println(output)
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

	return nil
}

func (s Service) Stop() ([]byte, error) {
	fmt.Println("Stopping", s.ConfigPath(), "...")
	return exec.Command("docker-compose", "-f", s.ConfigPath(), "stop").CombinedOutput()
}

func (s Service) getExposedURL(serviceName string, port string) (string, error) {
	output, err := exec.Command("docker-compose", "-f", s.ConfigPath(), "port", serviceName, port).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\n"), nil
}

func direct(prefix string, target *url.URL) func(req *http.Request) {
	regex := regexp.MustCompile(`^` + prefix)
	return func(req *http.Request) {
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = regex.ReplaceAllString(singleJoiningSlash(target.Path, req.URL.Path), "")
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
