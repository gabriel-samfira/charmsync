package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type BzrScm struct {
	name     string
	url      string
	workDir  string
	exePath  string
	revision string
}

func NewBzrScm(name, url, workDir, revision string) (*BzrScm, error) {
	exePath, err := exec.LookPath("bzr")
	if err != nil {
		return nil, err
	}
	return &BzrScm{
		workDir:  workDir,
		name:     name,
		url:      url,
		exePath:  exePath,
		revision: revision,
	}, nil
}

func (s *BzrScm) runBzrCommand(command ...string) error {
	bzrCommands := command
	bzrCommands = append(bzrCommands, s.FullPath())

	cmd := exec.Command(s.exePath, bzrCommands...)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (s *BzrScm) FullPath() string {
	return filepath.Join(s.workDir, s.name)
}

func (s *BzrScm) exists() bool {
	if _, err := os.Stat(s.FullPath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (s *BzrScm) initWorkdir() error {
	if _, err := os.Stat(s.workDir); os.IsNotExist(err) {
		err := os.MkdirAll(s.workDir, 00755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *BzrScm) gotoRevision(revision string) error {
	var err error = nil
	if revision != "" {
		err = s.runBzrCommand("revert", "-r", revision)
	} else {
		err = s.runBzrCommand("revert")
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *BzrScm) clone() error {
	if s.exists() {
		return fmt.Errorf("Repository %s already exists", s.FullPath())
	}
	cmd := exec.Command(s.exePath, "branch", s.url, s.FullPath())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error running git clone: %s --> %s", err, out)
	}
	return nil
}

func (s *BzrScm) update(revision string) error {
	fmt.Println("Updating repository: " + s.FullPath())
	exists := s.exists()
	if !exists {
		err := s.clone()
		if err != nil {
			return err
		}
	} else {
		isClean, err := s.isClean()
		if err != nil {
			return err
		}
		if !isClean {
			return fmt.Errorf("%s is not clean. Please commit local changes", s.FullPath())
		}
	}
	err := s.gotoRevision(revision)
	if err != nil {
		return err
	}
	return nil
}

func (s *BzrScm) isClean() (bool, error) {
	exists := s.exists()
	if !exists {
		return false, fmt.Errorf("Repository %s does not exists", s.FullPath())
	}
	cmd := exec.Command(s.exePath, "status", s.FullPath())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	if string(out) != "" {
		return false, nil
	}
	return true, nil
}

func (s *BzrScm) Name() string {
	return s.name
}

func (s *BzrScm) Update() error {
	err := s.initWorkdir()
	if err != nil {
		return err
	}
	err = s.update(s.revision)
	return err
}
