package main

import (
	"os/exec"
	"strings"
)

// TODO: this assume `etcd` and `etcdctl` is in APTH

func InitEtcd() error {
	cmd := exec.Command("etcd")
	return cmd.Start()
}

func AddToEtcd(s Service) {
	cmd := exec.Command("etcdctl", "set", "spacer.services."+s.Name, s.ExposedURLs["web"].String())
	_, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
}

func GetFromEtcd(s Service) string {
	cmd := exec.Command("etcdctl", "get", "spacer.services."+s.Name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(out))
}
