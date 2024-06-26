package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/phdavis1027/go-irodsclient/fs"
	"github.com/phdavis1027/go-irodsclient/irods/types"

	log "github.com/sirupsen/logrus"
)

func main() {
	logger := log.WithFields(log.Fields{
		"package":  "main",
		"function": "main",
	})

	recurse := false
	// Parse cli parameters
	flag.BoolVar(&recurse, "p", false, "create parent directories if not exist")
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Give an iRODS path!\n")
		os.Exit(1)
	}

	inputPath := args[0]

	// Read account configuration from YAML file
	yaml, err := os.ReadFile("account.yml")
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	account, err := types.CreateIRODSAccountFromYAML(yaml)
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	logger.Debugf("Account : %v", account.GetRedacted())

	// Create a file system
	appName := "make_dir"
	filesystem, err := fs.NewFileSystemWithDefault(account, appName)
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	defer filesystem.Release()

	err = filesystem.MakeDir(inputPath, recurse)
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	if filesystem.ExistsDir(inputPath) {
		fmt.Printf("Successfully made dir %s\n", inputPath)
	} else {
		fmt.Printf("Could not make dir %s\n", inputPath)
	}
}
