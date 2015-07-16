package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var cmdWrite = &Command{
	UsageLine: "write [-r] [-p] [-hash] [-file]",
	Short:     "writes imports of the package",
	Long: `writes imports of the package

-r	writes imports recursively, do note that the dependent 
	repositories should exist
-p	specify the directory of the package, by default it is "."
-hash	prints out the commit hash of each repository
-file	file to write to, defaults to Godeps
`,
}

func init() {
	cmdWrite.Run = runWrite // break init loop
}

var (
	writeDir       = cmdWrite.Flag.String("p", ".", "path of the go package")
	writeFile      = cmdWrite.Flag.String("file", "Godeps", "file to write to")
	writeRecursive = cmdWrite.Flag.Bool("r", true, "recursively write imports")
	writeHash      = cmdWrite.Flag.Bool("hash", true, "print out the commit hash")
)

func runWrite(cmd *Command, args []string) {
	file, err := os.OpenFile(
		filepath.Join(*writeDir, *writeFile),
		os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		elog.Println(err)
	}
	defer file.Close()

	for _, imp := range list(*listDir, *listRecursive, *listHash) {
		fmt.Fprintf(file, "%s %s\n", imp.Package, imp.Hash)
	}
}
