package main

import (
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func (s Service) Clone() error {
	output, err := exec.Command("git", "clone", "--depth=1", s.RepoCloneURL(), s.LocalRepoPath()).CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "already exists and is not an empty directory") {
			return ErrLocalPathAlreadyExists
		} else {
			return err
		}
	}

	return nil
}

func (s Service) RepoCloneURL() string {
	return "git@github.com:" + s.Repo.username + "/" + s.Repo.reponame + ".git"
}
