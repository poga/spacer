package main

import (
	"errors"
	"fmt"
	"os"
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
	os.MkdirAll(s.Path, os.ModePerm)
	fmt.Println(s.LocalPath, s.RemotePath, "->", s.Path)

	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		return nil, ErrLocalPathAlreadyExists
	}

	if s.RemotePath != "" {
		return Exec("git", "clone", "--depth=1", s.RemotePath, s.Path+s.Name)
	} else if s.LocalPath != "" {
		return Exec("cp", "-r", s.LocalPath, s.Path)
	}

	return nil, errors.New("Unknown dependency format " + s.Name)
}

func (s *Service) VersionIdentifier() string {
	if s.Commit != "" {
		return s.Commit
	}

	if s.Tag != "" {
		return s.Tag
	}

	return ""
}
