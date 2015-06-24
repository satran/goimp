package main

import (
	"errors"
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

type errPkgNotFound struct {
	string
}

func (e *errPkgNotFound) Error() string {
	path := strings.Replace(e.string, gopathPrefix, "", -1)
	path = strings.Trim(path, "/")
	return fmt.Sprintf("'%s' not found in %s", path, gopathPrefix)
}

func freeze(dir string, recursive, write, list, hash bool) {
	var outfile *os.File
	var err error
	if write {
		outfile, err = os.OpenFile(filepath.Join(dir, depsFile),
			os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	pkgGoPath, err := getPkgGoPath(dir)

	imports, err := getImports(dir, recursive, nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, imp := range imports {
		var line string
		if hash {
			commit, err := getCommit(imp, gopathPrefix)
			if err != nil {
				log.Printf(
					"couldn't get commit hash for %s\nerr:%s",
					imp, err,
				)
			}
			line = fmt.Sprintf("%s %s\n", imp, commit)
		} else {
			line = fmt.Sprintf("%s\n", imp)
		}

		// Ignore the packages own subpackages
		if strings.HasPrefix(imp, pkgGoPath) {
			continue
		}

		if list {
			fmt.Print(line)
		}
		if !write || outfile == nil {
			continue
		}
		_, err = outfile.WriteString(line)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getPkgGoPath(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if gopathPrefix == "" {
		return "", errors.New("GOPATH must be set for recursive option")
	}
	return strings.Trim(strings.TrimLeft(abs, gopathPrefix), "/"), nil
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

	if gopathPrefix == "" {
		return nil, errors.New("GOPATH must be set for recursive option")
	}

	for _, imp := range imports.Export() {
		if initial != nil && initial.Contains(imp) {
			continue
		}
		path := filepath.Join(gopathPrefix, imp)
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
