// +build !windows

package sync

import (
	"os"
)

func isSymlink(path string) (bool, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	isLink := fi.Mode()&os.ModeSymlink == os.ModeSymlink
	if isLink {
		return true, nil
	}
	return false, nil
}
