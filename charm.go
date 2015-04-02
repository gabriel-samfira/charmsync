package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/gabriel-samfira/charmsync/sync"
)

type charmBase struct {
	Scm      Scm
	Url      string
	Revision string
	Name     string

	workDir    string
	scmHandler ScmHandler
}

func (s *charmBase) WorkDir() string {
	return s.workDir
}

func (s *charmBase) InitScmHandler(workdir string) error {
	s.workDir = workdir
	if s.Name == "" || s.workDir == "" || s.Url == "" {
		return fmt.Errorf(
			"Missing required value: %v %v %v", s.Name, s.workDir, s.Url)
	}
	scmHandler, err := s.Scm.NewHandler(s.Name, s.Url, s.workDir, s.Revision)
	if err != nil {
		return err
	}
	s.scmHandler = scmHandler
	return nil
}

type Dependency struct {
	charmBase
	Resources   []string
	Destination string
}

type Charm struct {
	charmBase
	Upstream     string
	Dependencies []*Dependency
}

func NewCharmHandler(workdir string) (*Charm, error) {
	manifest := filepath.Join(workdir, ManifestFile)
	contents, err := ioutil.ReadFile(manifest)

	if err != nil {
		return nil, err
	}

	charm := &Charm{}
	err = json.Unmarshal(contents, charm)

	if err != nil {
		return nil, err
	}

	err = charm.InitScmHandler(workdir)
	if err != nil {
		return nil, err
	}

	for _, v := range charm.Dependencies {
		err := v.InitScmHandler(charm.DepsDir())
		if err != nil {
			return nil, err
		}
	}
	return charm, nil
}

func (c *Charm) FetchDependencies() error {
	fmt.Println("Fetching dependencies")
	for _, v := range c.Dependencies {
		err := v.scmHandler.Update()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Charm) FetchUpstream() error {
	var upstream Scm = Bazaar
	// Always fetch latest revision
	handler, err := upstream.NewHandler(c.Name, c.Upstream, c.StagingDir(), "")
	if err != nil {
		return err
	}
	fmt.Println("Fetching upstream repository")
	err = handler.Update()
	if err != nil {
		return err
	}
	return nil
}

func (c *Charm) FetchRepo() error {
	fmt.Println("Fetching development repository")
	err := c.scmHandler.Update()
	if err != nil {
		return err
	}
	return nil
}

func (c *Charm) FetchAll() error {
	err := c.FetchUpstream()
	if err != nil {
		return err
	}
	err = c.FetchRepo()
	if err != nil {
		return err
	}
	err = c.FetchDependencies()
	if err != nil {
		return err
	}
	return nil
}

func (c *Charm) SyncWithDevelopment() error {
	devRepo := filepath.Join(c.workDir, c.Name)
	upstreamRepo := filepath.Join(c.StagingDir(), c.Name)
	excludes := []string{
		fmt.Sprintf(".*\\.%s.*", c.Scm),
		".*\\.bzr.*",
	}
	s, err := sync.NewSync(devRepo, upstreamRepo, excludes, true)
	if err != nil {
		return err
	}
	err = s.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c *Charm) SyncDependencies() error {
	dstPath := filepath.Join(c.StagingDir(), c.Name)
	for _, v := range c.Dependencies {
		fullPath := v.scmHandler.FullPath()
		excludes := []string{
			fmt.Sprintf(".*\\.%s.*", v.Scm),
			".*\\.bzr.*",
		}
		destination := filepath.Join(dstPath, v.Destination)
		for _, j := range v.Resources {
			srcPath := filepath.Join(fullPath, j)
			dstPath := filepath.Join(destination, j)
			fmt.Printf("%v %v\r\n", srcPath, dstPath)
			s, err := sync.NewSync(srcPath, dstPath, excludes, true)
			if err != nil {
				return err
			}
			err = s.Run()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Charm) SyncAll() error {
	err := c.SyncWithDevelopment()
	if err != nil {
		return err
	}
	err = c.SyncDependencies()
	if err != nil {
		return err
	}
	return nil
}

func (c *Charm) DepsDir() string {
	deps := filepath.Join(c.workDir, "dependencies")
	return deps
}

func (c *Charm) StagingDir() string {
	staging := filepath.Join(c.workDir, "staging")
	return staging
}
