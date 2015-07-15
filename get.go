package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/satran/goimp/vcs"
)

var cmdGet = &Command{
	UsageLine: "get [-file] [-reset]",
	Short:     "gets imports of the package",
	Long: `gets imports of the package

-p	specify the directory of the package, by default it is .
-file	file to get commits from, defaults to Godeps
-reset	fetches the lastest code in the master branch
`,
}

func init() {
	cmdGet.Run = runGet // break init loop
}

var (
	getDir   = cmdGet.Flag.String("p", ".", "path of the go package")
	getFile  = cmdGet.Flag.String("file", "Godeps", "file to get to")
	getReset = cmdGet.Flag.Bool("reset", false, "fetches the latest code in the master branch")
)

func runGet(cmd *Command, args []string) {
	imports := getImportsFromFile(*getDir, *getFile)
	done := make(chan struct{})
	for _, imp := range imports {
		if *getReset {
			imp.Hash = ""
		}
		go get(imp, done)
	}
	for _ = range imports {
		<-done
	}
}

func getImportsFromFile(dir, file string) []Import {
	path := filepath.Join(dir, file)

	content, err := ioutil.ReadFile(path)
	if err != nil {
		elog.Fatalf("error reading deps file: %s", err)
	}

	var ret []Import
	lines := strings.Split(strings.Trim(string(content), "\n"), "\n")
	for _, line := range lines {
		splits := strings.Split(line, " ")
		pkg := splits[0]
		var hash string
		if len(splits) > 1 {
			hash = splits[1]
		}
		ret = append(ret, Import{pkg, hash})
	}
	return ret
}

func get(imp Import, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()
	vcspath := filepath.Join(goPathSrc, imp.Package)
	if !exists(vcspath) {
		cmd := exec.Command("go", "get", imp.Package)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			elog.Print(err)
			return
		}
	}
	v, err := vcs.New(vcspath, goPathSrc)
	if err != nil {
		elog.Print(err)
	}
	if imp.Hash != "" {
		err = v.Checkout(imp.Hash)
		if err != nil {
			// try fetching for the latest commit
			err = v.Fetch()
			if err != nil {
				elog.Print(err)
				return
			}
			err = v.Checkout(imp.Hash)
			if err != nil {
				elog.Print(err)
				return
			}
		}
		return
	}
	if err := v.Latest(); err != nil {
		elog.Println(err)
	}
}

func exists(dir string) bool {
	dir = filepath.Clean(dir)
	_, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return true
}
