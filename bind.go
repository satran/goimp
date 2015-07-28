package main

import (
	"fmt"
	"reflect"
	"time"
)

var cmdBind = &Command{
	UsageLine: "bind [-file] [-p]",
	Short:     "binds Godeps file to imports",
	Long: `watches for Changes in the Godeps file or in each packges, updating either when detected

-p	specify the directory of the package, by default it is "."
-file	file to get commits from, defaults to Godeps
`,
}

type Bind struct{}

func init() {
	cmdBind.Run = runBind // break init loop
}

var (
	getBindDir   = cmdBind.Flag.String("p", ".", "path of the go package")
	getBindFile  = cmdBind.Flag.String("file", "Godeps", "file to get to")
	getBindReset = cmdBind.Flag.Bool("reset", false, "fetches the latest code in the master branch")
)

func importsToMap(arg []Import) map[string]string {
	ret := make(map[string]string)
	for _, imp := range arg {
		ret[imp.Package] = imp.Hash
	}
	return ret
}

func runBind(cmd *Command, args []string) {
	var readImports map[string]string
	checkReadImports := func() bool {
		newImports := importsToMap(getImportsFromFile(*getBindDir, *getBindFile))
		if !reflect.DeepEqual(readImports, newImports) {
			readImports = newImports
			return true
		}
		return false
	}

	var writeImports map[string]string
	checkWriteImports := func() bool {
		newImports := importsToMap(list(".", true, true))
		if !reflect.DeepEqual(writeImports, newImports) {
			return true
		}
		return false
	}

	for {
		readImports = importsToMap(getImportsFromFile(*getBindDir, *getBindFile))
		writeImports = importsToMap(list(".", true, true))

		for {
			time.Sleep(time.Second)
			if checkReadImports() {
				fmt.Println("getting...")
				runGet(&Command{}, []string{})
				break
			}
			if checkWriteImports() {
				fmt.Println("writing...")
				runWrite(&Command{}, []string{})
				break
			}
		}
	}
}
