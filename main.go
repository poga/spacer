package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
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
		/*
			dir := strings.Split(l, "/")
			if len(dir) < 2 {
				continue
			}
				url := "git@github.com:" + dir[0] + "/" + dir[1] + ".git"
				os.Mkdir("services", 0777)
				to := fmt.Sprintf("services/%s", dir[1])
				output, err := exec.Command("git", "clone", url, to).CombinedOutput()
				if err != nil {
					log.Panic(err)
				}
				fmt.Println(string(output))
		*/
		s := NewService(l)
		services = append(services, s)

		// docker-compose build && docker-compose up
		dcb, err := s.Build()
		if err != nil {
			log.Panic(err)
		}
		fmt.Println(string(dcb))

		s.Start()
		/*

			time.Sleep(3 * time.Second)

			dcs, err := s.Stop()
			if err != nil {
				log.Panic(err)
			}
			fmt.Println(string(dcs))

		*/
	}

	fmt.Printf("%v\n", services)
	fmt.Printf("%v\n", exposedURLs)

	// setup a proxy for each service
	for _, s := range services {
		// TODO not just web
		prefix := "/" + s.Name
		fmt.Printf("Proxying %s to %s\n", prefix, s.ExposedURLs["web"])
		proxy := httputil.NewSingleHostReverseProxy(s.ExposedURLs["web"])
		proxy.Director = direct(prefix, s.ExposedURLs["web"])
		http.HandleFunc(prefix, proxy.ServeHTTP)
	}

	http.ListenAndServe(":9064", nil)
}

func direct(prefix string, target *url.URL) func(req *http.Request) {
	regex := regexp.MustCompile(`^` + prefix)
	return func(req *http.Request) {
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = regex.ReplaceAllString(singleJoiningSlash(target.Path, req.URL.Path), "")
		fmt.Printf("redirected: %s %s %s\n", req.URL.Scheme, req.URL.Host, req.URL.Path)
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
