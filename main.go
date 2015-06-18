package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	gopathPrefix string
	depsFile     string
)

func init() {
	log.SetOutput(os.Stderr)
	path := os.Getenv("GOPATH")
	if path == "" {
		log.Fatal("GOPATH not set")
	}
	gopath := strings.Split(path, ":")[0]
	os.Setenv("GOPATH", gopath)
	gopathPrefix = filepath.Join(gopath, "src")
}

func main() {
	debug := flag.Bool("debug", false, "error logs are prefixed with file name")
	write := flag.Bool("w", false, "writes dependency to deps file")
	list := flag.Bool("l", false, "lists dependency to stdout")
	hash := flag.Bool("hash", true, "adds commit hash when writing or listing")
	recursive := flag.Bool("r", false, "finds imports recursively")
	dir := flag.String("dir", ".", "project source directory")
	flag.StringVar(&depsFile, "deps", "Godeps", "the dependency file")
	flag.Parse()
	if *debug {
		log.SetFlags(log.Lshortfile)
	} else {
		log.SetFlags(0)
	}

	if *write || *list {
		freeze(*dir, *recursive, *write, *list, *hash)
	} else {
		get(*dir)
	}
}
