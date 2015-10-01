package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/satran/goimp/vcs"
)

var cmdGet = &Command{
	UsageLine: "get [-file] [-reset] [-p] [package_url commit]",
	Short:     "gets imports of the package",
	Long: `gets imports of the package

-p	specify the directory of the package, by default it is "."
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
	var imports []Import

	if len(args) > 0 {
		pkg := args[0]
		hash := "master"
		if len(args) == 2 {
			if *getReset {
				elog.Fatal("reset argument conflicts when specifying the commit")
			}
			hash = args[1]
		}
		imports = []Import{
			{Package: pkg, Hash: hash},
		}
	} else {
		imports = getImportsFromFile(*getDir, *getFile)
	}

	var wg sync.WaitGroup
	wg.Add(len(imports))
	for _, imp := range imports {
		go func(imp Import) {
			getDependencies(imp)
			wg.Done()
		}(imp)
	}
	wg.Wait()

	wg.Add(len(imports))
	for _, imp := range imports {
		if *getReset {
			imp.Hash = ""
		}
		go func(imp Import) {
			get(imp)
			wg.Done()
		}(imp)
	}
	wg.Wait()
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
		splits := strings.Fields(line)
		pkg := splits[0]
		var hash string
		if len(splits) > 1 {
			hash = splits[1]
		}
		ret = append(ret, Import{pkg, hash})
	}
	return ret
}

func getDependencies(imp Import) {
	vcspath := filepath.Join(goPathSrc,
		strings.TrimRight(imp.Package, "/..."))
	if !exists(vcspath) {
		cmd := exec.Command("go", "get", "-d", imp.Package)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			elog.Print(err)
			return
		}
	}
}

func get(imp Import) {
	vcspath := filepath.Join(goPathSrc,
		strings.TrimRight(imp.Package, "/..."))
	v, err := vcs.New(vcspath, goPathSrc)
	if err != nil {
		elog.Print(err)
		return
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
				elog.Printf("error checkout out %s: %s", imp.Package, err)
				return
			}
		}
		return
	}
	if err := v.Latest(); err != nil {
		elog.Printf("error trying to set %s to latest: %s", imp.Package, err)
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
