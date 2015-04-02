package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type GitScm struct {
	name     string
	url      string
	workDir  string
	gitPath  string
	revision string
}

func NewGitScm(name, url, workDir, revision string) (*GitScm, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	return &GitScm{
		workDir:  workDir,
		name:     name,
		url:      url,
		gitPath:  gitPath,
		revision: revision,
	}, nil
}

func (s *GitScm) Name() string {
	return s.name
}

func (s *GitScm) FullPath() string {
	return filepath.Join(s.workDir, s.name)
}

func (s *GitScm) exists() bool {
	if _, err := os.Stat(s.FullPath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (s *GitScm) clone() error {
	if s.exists() {
		return fmt.Errorf("Repository %s already exists", s.FullPath())
	}
	cmd := exec.Command(s.gitPath, "clone", s.url, s.FullPath())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error running git clone: %s --> %s", err, out)
	}
	return nil
}

func (s *GitScm) runGitCommand(command ...string) error {
	gitCommands := []string{
		"-C",
		s.FullPath(),
	}
	gitCommands = append(gitCommands, command...)

	cmd := exec.Command(s.gitPath, gitCommands...)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (s *GitScm) gotoRevision(revision string) error {
	err := s.runGitCommand("checkout", "master")
	if err != nil {
		return err
	}
	err = s.runGitCommand("pull")
	if err != nil {
		return err
	}
	if revision != "" {
		err = s.runGitCommand("checkout", revision)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *GitScm) update(revision string) error {
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

func (s *GitScm) isClean() (bool, error) {
	exists := s.exists()
	if !exists {
		return false, fmt.Errorf("Repository %s does not exists", s.FullPath())
	}
	cmd := exec.Command(s.gitPath, "-C", s.FullPath(), "status", "-s")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	if string(out) != "" {
		return false, nil
	}
	return true, nil
}

func (s *GitScm) initWorkdir() error {
	if _, err := os.Stat(s.workDir); os.IsNotExist(err) {
		err := os.MkdirAll(s.workDir, 00755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *GitScm) Update() error {
	err := s.initWorkdir()
	if err != nil {
		return err
	}
	err = s.update(s.revision)
	return err
}
