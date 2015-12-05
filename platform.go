package main

type Platform interface {
	Build(Service) ([]byte, error)
	Start(Service) error
	Stop(Service) ([]byte, error)
	ConfigPath(Service) string
}
