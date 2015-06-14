package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	goPathPrefix string
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		splits := strings.Split(gopath, ":")
		// Use only the first of the gopath
		goPathPrefix = filepath.Join(splits[0], "src")
	}
}

type errPkgNotFound struct {
	string
}

func (e *errPkgNotFound) Error() string {
	path := strings.TrimLeft(e.string, goPathPrefix)
	return fmt.Sprintf("'%s' not found", path)
}

func main() {
	recursive := flag.Bool("r", false, "finds imports recursively")
	directory := flag.String("dir", ".", "package directory for which the imports must be found")
	commit := flag.Bool("hash", false, "whether to return the commit hash")
	flag.Parse()

	imports, err := getImports(*directory, *recursive, nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, imp := range imports {
		if !*commit {
			fmt.Println(imp)
			continue
		}
		commit, err := getCommit(imp)
		if err != nil {
			log.Printf("couldn't get commit hash for %s\nerr:%s", imp, err)
		}
		fmt.Println(imp, commit)
	}
}

func getImports(dir string, recursive bool, initial *set) ([]string, error) {
	imports := newSet()
	files, err := parseDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		imports.Extend(getFileImports(file))
	}
	if !recursive {
		return imports.Export(), nil
	}

	if goPathPrefix == "" {
		return nil, errors.New("GOPATH must be set for recursive option")
	}

	for _, imp := range imports.Export() {
		if initial != nil && initial.Contains(imp) {
			continue
		}
		path := filepath.Join(goPathPrefix, imp)
		pathImports, err := getImports(path, recursive, imports)
		switch err.(type) {
		case *errPkgNotFound:
			log.Print(err)
		case nil:
		default:
			return nil, err
		}
		imports.Extend(pathImports)
	}
	return imports.Export(), nil
}

// parseDir parses the package in the given directory and returns it.
func parseDir(directory string) ([]*ast.File, error) {
	dirFiles, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, &errPkgNotFound{directory}
	}

	var files []*ast.File
	fs := token.NewFileSet()
	for _, fd := range dirFiles {
		if fd.IsDir() {
			continue
		}
		name := fd.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if directory != "." {
			name = filepath.Join(directory, name)
		}
		f, err := parser.ParseFile(fs, name, nil, 0)
		if err != nil {
			log.Printf("ignoring unparsable file %q: %s", name, err)
			continue
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("%s: no buildable Go files", directory)
	}
	return files, nil
}

func getFileImports(file *ast.File) []string {
	imports := newSet()
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if !isStdLib(path) {
			imports.Add(path)
		}
	}
	return imports.Export()
}
