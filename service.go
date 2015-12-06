package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type Service struct {
	RemotePath string // 從哪裡 clone
	LocalPath  string // 或是從哪裡 copy
	Name       string

	Commit string // 想要的 commit hash
	Tag    string // 或是想要的 version tag

	Path string // 最後 service 被存到哪了
}

var ErrLocalPathAlreadyExists = errors.New("local path already exists")

func (s *Service) Fetch() ([]byte, error) {
	var output []byte
	var err error

	os.MkdirAll(s.Path, os.ModePerm)
	fmt.Println(s.LocalPath, "->", s.Path)

	if s.RemotePath != "" {
		output, err = Exec("git", "clone", "--depth=1", s.RemotePath, s.Path)
		if err != nil {
			log.Println(string(output))
			if strings.Contains(string(output), "already exists and is not an empty directory") {
				return output, ErrLocalPathAlreadyExists
			}
		}
	} else if s.LocalPath != "" {
		output, err = Exec("cp", "-r", s.LocalPath, s.Path)
	}
	return output, err
}
