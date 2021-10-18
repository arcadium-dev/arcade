package main

import (
	"fmt"
	l "log"
	"os"

	"arcadium.dev/build"
	"arcadium.dev/log"
)

var (
	version string
	branch  string
	shasum  string
	date    string
)

func main() {
	info := build.Info(version, branch, shasum, date)
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(info)
		os.Exit(0)
	}

	logger, err := log.New()
	if err != nil {
		l.Fatalf("Exiting: %s", err)
	}
	logger.Info("msg", "starting")
	logger.Info(info.Fields()...)
}
