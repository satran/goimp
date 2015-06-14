package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func get(dir string) {
	content, err := ioutil.ReadFile(depsFile)
	if err != nil {
		log.Fatalf("error reading deps file: %s", err)
	}

	deps := strings.Split(strings.Trim(string(content), "\n"), "\n")
	for _, dep := range deps {
		splits := strings.Split(dep, " ")
		imp := splits[0]
		var hash string
		if len(splits) > 1 {
			hash = splits[1]
		}
		if exists(filepath.Join(gopathPrefix, imp)) == nil {
			log.Printf("fetching %s...", imp)
			err = Fetch(imp, gopathPrefix)
		} else {
			log.Printf("go get %s...", imp)
			err = Execute("", "go", "get", "-u", imp)
		}
		if err != nil {
			log.Printf("%s", err)
			continue
		}
		if hash == "" {
			continue
		}
		log.Printf("checkout to %s...", hash)
		err = Checkout(imp, gopathPrefix, hash)
		if err != nil {
			log.Printf("%s", err)
			continue
		}
	}
}

func exists(dir string) error {
	dir = filepath.Clean(dir)
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}
	return nil
}
