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
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	builtins     = newSet()
	goPathPrefix string
	logE         = log.New(os.Stderr, "", 0)

	recursive = flag.Bool("r", false, "finds imports recursively")
	directory = flag.String("dir", ".",
		"package directory for which the imports must be found")
	commit = flag.Bool("hash", false, "whether to return the commit hash")
)

func init() {
	stdlib := []string{
		"archive", "bufio", "builtin", "bytes", "compress", "container",
		"crypto", "database", "debug", "encoding", "errors", "expvar",
		"flag", "fmt", "go", "hash", "html", "image", "index", "io",
		"log", "math", "mime", "net", "os", "path", "reflect", "regexp",
		"runtime", "sort", "strconv", "strings", "sync", "syscall",
		"testing", "text", "time", "unicode", "unsafe",
	}
	for _, path := range stdlib {
		builtins.Add(path)
	}
	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		goPathPrefix = filepath.Join(gopath, "src")
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
	flag.Parse()
	imports, err := getImports(*directory, *recursive, nil)
	if err != nil {
		logE.Fatal(err)
	}
	for _, imp := range imports {
		if *commit {
			commit, err := getCommit(imp)
			if err != nil {
				logE.Printf(
					"couldn't get commit hash for %s\nerr:%s",
					imp, err,
				)
			}
			fmt.Println(imp, commit)
		} else {
			fmt.Println(imp)
		}
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
			logE.Print(err)
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
			logE.Printf("ignoring unparsable file %q: %s", name, err)
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

func isStdLib(path string) bool {
	p := strings.SplitN(path, "/", 2)[0]
	return builtins.Contains(p)
}

type vcsCmd struct {
	name   string
	cmd    string
	commit string
}

var vcsList = []*vcsCmd{
	{
		name:   "Git",
		cmd:    "git",
		commit: "git rev-parse HEAD",
	},
	{
		name:   "Mercurial",
		cmd:    "hg",
		commit: "hg id -i",
	},
}

// vcsForDir inspects dir and its parents to determine the
// version control system and code repository to use.
// On return, root is the import path
// corresponding to the root of the repository
// (thus root is a prefix of importPath).
func vcsForDir(dir string) (vcs *vcsCmd, root string, err error) {
	// Clean and double-check that dir is in (a subdirectory of) srcRoot.
	dir = filepath.Clean(dir)
	srcRoot := filepath.Clean(goPathPrefix)
	if len(dir) <= len(srcRoot) || dir[len(srcRoot)] != filepath.Separator {
		return nil, "", fmt.Errorf(
			"directory %q is outside source root %q", dir, srcRoot)
	}

	origDir := dir
	for len(dir) > len(srcRoot) {
		for _, vcs := range vcsList {
			fi, err := os.Stat(filepath.Join(dir, "."+vcs.cmd))
			if err == nil && fi.IsDir() {
				return vcs, dir, nil
			}
		}

		// Move to parent.
		ndir := filepath.Dir(dir)
		if len(ndir) >= len(dir) {
			// Shouldn't happen, but just in case, stop.
			break
		}
		dir = ndir
	}

	return nil, "", fmt.Errorf(
		"directory %q is not using a known version control system",
		origDir,
	)
}

func getCommit(imp string) (string, error) {
	vcs, root, err := vcsForDir(filepath.Join(goPathPrefix, imp))
	if err != nil {
		return "", err
	}
	if vcs.commit == "" {
		return "", fmt.Errorf("%s is not yet supported", vcs.name)
	}
	args := strings.Split(vcs.commit, " ")
	name := args[0]
	if len(args) > 2 {
		args = args[1:]
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\n"), nil
}
