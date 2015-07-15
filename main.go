package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	goPathSrc string
	depsFile  string
)

var (
	elog = log.New(os.Stderr, "", log.Lshortfile)
)

var commands = []*Command{
	cmdList,
	cmdWrite,
	cmdGet,
}

func init() {
	log.SetFlags(0)
	path := os.Getenv("GOPATH")
	if path == "" {
		log.Fatal("GOPATH not set")
	}
	gopath := strings.Split(path, ":")[0]
	os.Setenv("GOPATH", gopath)
	goPathSrc = filepath.Join(gopath, "src")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	if args[0] == "help" {
		help(args[1:])
		return
	}
	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			cmd.Run(cmd, args)
			return
		}
	}
	elog.Printf("goimp: unknown subcommand %q\nRun goimp help' for usage.\n", args[0])
}
