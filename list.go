package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/satran/goimp/vcs"
)

type Import struct {
	Package string
	Hash    string
}

type errPkgNotFound struct {
	string
}

func (e *errPkgNotFound) Error() string {
	path := strings.Replace(e.string, goPathSrc, "", -1)
	path = strings.Trim(path, "/")
	return fmt.Sprintf("'%s' not found in %s", path, goPathSrc)
}

var cmdList = &Command{
	UsageLine: "list [-r] [-p] [-hash]",
	Short:     "lists imports of the package",
	Long: `lists imports of the package

-r	lists imports recursively, do note that the dependent 
	repositories should exist
-p	specify the directory of the package, by default it is "."
-hash	prints out the commit hash of each repository
`,
}

func init() {
	cmdList.Run = runList // break init loop
}

var (
	listDir       = cmdList.Flag.String("p", ".", "path of the go package")
	listRecursive = cmdList.Flag.Bool("r", false, "recursively list imports")
	listHash      = cmdList.Flag.Bool("hash", false, "print out the commit hash")
)

func runList(cmd *Command, args []string) {
	for _, imp := range list(*listDir, *listRecursive, *listHash) {
		fmt.Printf("%s %s\n", imp.Package, imp.Hash)
	}
}

func list(dir string, recursive, hash bool) []Import {
	imports, err := getPackageImports(*listDir, *listRecursive, nil)
	if err != nil {
		elog.Fatal(err)
	}
	imports = purgeSubPackages(*listDir, imports)
	sorted := sort.StringSlice(imports)
	sort.Sort(sorted)
	var ret []Import
	for _, pkg := range sorted {
		imp := Import{Package: pkg}
		if !*listHash {
			ret = append(ret, imp)
			continue
		}
		path := filepath.Join(goPathSrc, pkg)
		v, err := vcs.New(path, goPathSrc)
		if err != nil {
			elog.Println(err)
			continue
		}
		hash, err := v.CommitHash()
		if err != nil {
			elog.Println(err)
			continue
		}
		imp.Hash = hash
		ret = append(ret, imp)
	}
	return ret
}

func purgeSubPackages(pkg string, imports []string) []string {
	var ret []string
	for _, imp := range imports {
		if strings.HasPrefix(imp, pkgPath(pkg)) {
			continue
		}
		ret = append(ret, imp)
	}
	return ret
}

func pkgPath(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		elog.Fatal(err)
	}
	return strings.Trim(strings.TrimLeft(abs, goPathSrc), "/")
}

func getPackageImports(dir string, recursive bool, initial *set) ([]string, error) {
	imports := newSet()
	files, err := parseDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		imports.Extend(importsForFile(file))
	}
	if !recursive {
		return imports.Export(), nil
	}

	if goPathSrc == "" {
		return nil, errors.New("GOPATH must be set for recursive option")
	}

	for _, imp := range imports.Export() {
		if initial != nil && initial.Contains(imp) {
			continue
		}
		path := filepath.Join(goPathSrc, imp)
		pathImports, err := getPackageImports(path, recursive, imports)
		switch err.(type) {
		case *errPkgNotFound:
			elog.Print(err)
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
			elog.Printf("ignoring unparsable file %q: %s", name, err)
			continue
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("%s: no buildable Go files", directory)
	}
	return files, nil
}

func importsForFile(file *ast.File) []string {
	imports := newSet()
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if !isStdLib(path) {
			imports.Add(path)
		}
	}
	return imports.Export()
}
