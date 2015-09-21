package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

func main() {
	var services []Service
	var exposedURLs []string

	os.RemoveAll("./services")

	spacerfile, _ := ioutil.ReadFile("Spacerfile.example")
	lines := strings.Split(string(spacerfile), "\n")
	for _, l := range lines {
		fmt.Println(l)

		if len(l) == 0 {
			continue
		}

		// git clone path
		dir := strings.Split(l, "/")
		if len(dir) < 2 {
			continue
		}
		/*
			url := "git@github.com:" + dir[0] + "/" + dir[1] + ".git"
			os.Mkdir("services", 0777)
			to := fmt.Sprintf("services/%s", dir[1])
			output, err := exec.Command("git", "clone", url, to).CombinedOutput()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(output))
		*/
		s := Service{l, dir[1]}
		services = append(services, s)

		// docker-compose build && docker-compose up
		/*
			dcb, err := s.Build()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(dcb))

			s.Start()

			time.Sleep(3 * time.Second)

			dcs, err := s.Stop()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(dcs))

		*/
		// loop through docker-compose.yml and look for exposed ip, save it
		ymlBytes, err := ioutil.ReadFile(s.ConfigPath())
		if err != nil {
			log.Fatal(err)
		}

		m := make(map[interface{}]interface{})
		err = yaml.Unmarshal(ymlBytes, &m)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v\n", m)

		// if ports == "5000:5000", take the former port
		// if ports == "5000", ask "docker-compose port SERVICE_NAME 5000" to know the exposed port
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
				output, err := s.GetExposedURL(serviceName.(string), innerPort)
				if err != nil {
					log.Fatal(err)
				}
				exposedURLs = append(exposedURLs, output)

			}
		}

	}

	fmt.Printf("%v\n", services)
	fmt.Printf("%v\n", exposedURLs)

	// setup a proxy to all services
	proxy := httputil.NewSingleHostReverseProxy(nil)
	proxy.Director = func(r *http.Request) {
		for _, s := range services {
			if strings.HasPrefix(r.URL.Path, s.Name) {
				// TODO change URL's host to container url
			}
		}
	}
}
