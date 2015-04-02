package main

import (
	"fmt"
	"os"
)

const ManifestFile = "charmsync.json"

func main() {
	wrkDir, err := os.Getwd()
	if err != nil {
		msg := fmt.Errorf("Failed to get current working directory: %s", err)
		fmt.Println(msg)
		os.Exit(1)
	}

	charm, err := NewCharmHandler(wrkDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = charm.FetchAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = charm.SyncAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
