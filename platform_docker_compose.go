package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type DockerCompose struct {
	Host string
}

func NewDockerCompose(dockerHost string) Platform {
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

	return DockerCompose{host}
}

func (p DockerCompose) Build(s Service) error {
	_, err := exec.Command("docker-compose", "-f", p.ConfigPath(s), "build").CombinedOutput()
	return err
}

func (p DockerCompose) Stop(s Service) error {
	_, err := exec.Command("docker-compose", "-f", p.ConfigPath(s), "stop").CombinedOutput()
	return err
}

func (p DockerCompose) Start(s Service) error {
	err := exec.Command("docker-compose", "-f", p.ConfigPath(s), "up").Start()
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	// loop through docker-compose.yml and look for exposed ip, save it
	ymlBytes, err := ioutil.ReadFile(p.ConfigPath(s))
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
			urlStr, err := p.getExposedURLString(s, serviceName, innerPort)
			if err != nil {
				log.Panic(err)
			}
			u, err := url.Parse("http://" + urlStr)
			if err != nil {
				log.Panic(err)
			}
			if p.Host != "" {
				_, originPort, err := net.SplitHostPort(u.Host)
				if err != nil {
					return err
				}
				u.Host = p.Host + ":" + originPort
			}
			s.ExposedURLs[serviceName] = u
			AddToEtcd(s)
		}
	}

	return nil

}

func (p DockerCompose) getExposedURLString(s Service, serviceName string, port string) (string, error) {
	output, err := exec.Command("docker-compose", "-f", p.ConfigPath(s), "port", serviceName, port).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\n"), nil
}

func (p DockerCompose) ConfigPath(s Service) string {
	return s.Path + "/docker-compose.yml"
}
