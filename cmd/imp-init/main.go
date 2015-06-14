package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	vendorPath = ".vendor"
	sourceFile = "goimp.env"
)

var source = `export GOPATH=%s:$GOPATH`

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
}

func main() {
	repo := flag.String("repo", ".", "directory containing source files")
	vendor := flag.String("vendor", "", "directory to store the vendor repositories. (default: repo/.vendor)")
	flag.Parse()

	err := exists(*repo)
	if err != nil {
		log.Fatal(err)
	}

	vendorDir, err := initVendor(*vendor, *repo)
	if err != nil {
		log.Fatal(err)
	}
	abspath, err := filepath.Abs(vendorDir)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(sourceFile, []byte(fmt.Sprintf(source, abspath)), 0644)
	if err != nil {
		log.Fatal(err)
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

func initVendor(vendor, repo string) (string, error) {
	if vendor == "" {
		vendor = filepath.Join(repo, vendorPath)
	}
	vendor = filepath.Clean(vendor)
	vendorSrc := filepath.Join(vendor, "src")
	err := exists(vendorSrc)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		err = os.MkdirAll(vendorSrc, 0755)
		if err != nil {
			return "", err
		}
	}
	return vendor, nil
}

func initRepo(repo, vendor string) error {
	repo = filepath.Clean(repo)
	_, err := os.Stat(filepath.Clean(repo))
	if err != nil {
		return err
	}

	return nil
}
