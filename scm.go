package main

import (
	"fmt"
)

const (
	Git    Scm = "git"
	Bazaar     = "bzr"
)

type ScmHandler interface {
	Name() string
	Update() error
	FullPath() string
}

type Scm string
type Upstream string

func (s Scm) NewHandler(name, url, workDir, revision string) (ScmHandler, error) {
	switch s {
	case "git":
		return NewGitScm(name, url, workDir, revision)
	case "bzr":
		return NewBzrScm(name, url, workDir, revision)
	default:
		return nil, fmt.Errorf("Unsupported SCM system")
	}
}

func (s Upstream) NewHandler(name, url, workDir, revision string) (ScmHandler, error) {
	switch s {
	case "git":
		return NewGitScm(name, url, workDir, revision)
	default:
		return nil, fmt.Errorf("Unsupported SCM system")
	}
}
