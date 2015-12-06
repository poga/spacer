package main

import "net/url"

type Platform interface {
	ConfigPath(Service) string

	Build(Service) ([]byte, error)
	Start(Service) error
	Stop(Service) ([]byte, error)

	Running() map[string]*url.URL
}
