package main

type Platform interface {
	Build(Service) error
	Start(Service) error
	Stop(Service) error
	ConfigPath(Service) string
}
