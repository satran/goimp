package main

import (
	"strings"
)

var isStdLib func(path string) bool

func init() {
	builtins := newSet()
	for _, path := range []string{
		"C", "archive", "bufio", "builtin", "bytes", "compress", "container",
		"crypto", "database", "debug", "encoding", "errors", "expvar",
		"flag", "fmt", "go", "hash", "html", "image", "index", "io",
		"log", "math", "mime", "net", "os", "path", "reflect", "regexp",
		"runtime", "sort", "strconv", "strings", "sync", "syscall",
		"testing", "text", "time", "unicode", "unsafe"} {
		builtins.Add(path)
	}
	isStdLib = func(path string) bool {
		p := strings.SplitN(path, "/", 2)[0]
		return builtins.Contains(p)
	}
}
