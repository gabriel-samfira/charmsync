package sync

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gabriel-samfira/charmsync/util"
)

type info struct {
	checksum string
	isDir    bool
	size     int64
	mtime    time.Time
	isLink   bool
}

type Sync struct {
	src      string
	dst      string
	exclude  []string
	checksum bool

	dstMap    map[string]info
	srcMap    map[string]info
	changeMap map[string]info
}

func computeHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	hash := sha512.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	file.Close()
	h := hex.EncodeToString(hash.Sum(nil))
	return string(h), nil
}

func NewSync(src, dst string, exclude []string, checksum bool) (*Sync, error) {
	err := os.MkdirAll(dst, 00755)
	if err != nil {
		return nil, err
	}
	return &Sync{
		src:      src,
		dst:      dst,
		exclude:  exclude,
		checksum: checksum,

		dstMap:    map[string]info{},
		srcMap:    map[string]info{},
		changeMap: map[string]info{},
	}, nil
}

func buildRegex(excludes []string) string {
	return strings.Join(excludes, "|")
}

func shouldSkip(path, regex string) bool {
	if regex == "" {
		return false
	}
	if matched, err := regexp.MatchString(regex, path); err != nil {
		return false
	} else {
		if matched {
			return true
		}
	}
	return false
}

func getInfo(root, path string, f os.FileInfo, checksum bool) (info, string, error) {
	var sum string
	relpath, err := filepath.Rel(root, path)
	if err != nil {
		return info{}, "", err
	}
	isLink, err := isSymlink(path)
	if err != nil {
		return info{}, "", err
	}
	if checksum && !f.IsDir() && !isLink {
		sum, err = computeHash(path)
		if err != nil {
			return info{}, "", err
		}
	}
	inf := info{
		isDir:    f.IsDir(),
		size:     f.Size(),
		mtime:    f.ModTime(),
		isLink:   isLink,
		checksum: sum,
	}
	return inf, relpath, nil
}

func (s *Sync) buildMaps() error {
	excludesRe := buildRegex(s.exclude)
	srcFunc := func(path string, f os.FileInfo, err error) error {
		if shouldSkip(path, excludesRe) {
			return nil
		}
		inf, relpath, err := getInfo(s.src, path, f, s.checksum)
		if err != nil {
			return err
		}
		s.srcMap[relpath] = inf
		return nil
	}

	dstFunc := func(path string, f os.FileInfo, err error) error {
		if shouldSkip(path, excludesRe) {
			return nil
		}
		inf, relpath, err := getInfo(s.dst, path, f, s.checksum)
		if err != nil {
			return err
		}
		s.dstMap[relpath] = inf
		return nil
	}

	errChan := make(chan error, 1)

	go func() {
		err := filepath.Walk(s.src, srcFunc)
		errChan <- err
	}()

	go func() {
		err := filepath.Walk(s.dst, dstFunc)
		errChan <- err
	}()

	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			if err != nil {
				return err
			}
		case <-time.After(10 * time.Minute):
			return fmt.Errorf("Folder walk took longer then 10 minutes. What kind of charm is this?")
		}
	}

	return nil
}

func (s *Sync) compareMaps() {
	for k, v := range s.srcMap {
		if val, ok := s.dstMap[k]; ok {
			if val.isDir || val.isLink {
				delete(s.dstMap, k)
				delete(s.srcMap, k)
				continue
			}
			if s.checksum {
				if v.checksum != val.checksum {
					s.changeMap[k] = v
				}
			} else {
				if v.size != val.size {
					s.changeMap[k] = v
				}
			}
			delete(s.dstMap, k)
			delete(s.srcMap, k)
		}
	}
}

func (s *Sync) createBaseDir(root, path string) error {
	dirname := filepath.Dir(path)
	dstDir := filepath.Join(root, dirname)
	exists := util.PathExists(dstDir)
	if !exists {
		err := os.MkdirAll(dstDir, 00755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sync) copyToDestination(path string) error {
	err := s.createBaseDir(s.dst, path)
	if err != nil {
		return fmt.Errorf("Failed to create directory: %s", err)
	}
	srcFile := filepath.Join(s.src, path)
	dstFile := filepath.Join(s.dst, path)
	fmt.Printf("Copying %s to %s\r\n", srcFile, dstFile)
	err = util.CopyFile(srcFile, dstFile)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sync) Run() error {
	err := s.buildMaps()
	if err != nil {
		return err
	}

	s.compareMaps()
	// Remove inexistent files on source
	for k, _ := range s.dstMap {
		fullPath := filepath.Join(s.dst, k)
		fmt.Printf("Deleting file %s\r\n", fullPath)
		err := os.RemoveAll(fullPath)
		if err != nil {
			return err
		}
	}

	for k, v := range s.srcMap {
		if v.isDir {
			continue
		}
		err := s.copyToDestination(k)
		if err != nil {
			return err
		}
	}

	for k, v := range s.changeMap {
		if v.isDir {
			continue
		}
		err := s.copyToDestination(k)
		if err != nil {
			return err
		}
	}
	return nil
}
