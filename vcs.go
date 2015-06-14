package goimp

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type vcsCmd struct {
	name     string
	cmd      string
	commit   string
	checkout string
	fetch    string
}

var vcsList = []*vcsCmd{
	{
		name:     "Git",
		cmd:      "git",
		commit:   "git rev-parse HEAD",
		checkout: "git checkout",
		fetch:    "git fetch",
	},
	{
		name:     "Mercurial",
		cmd:      "hg",
		commit:   "hg id -i",
		checkout: "hg update",
		fetch:    "hg pull",
	},
}

// vcsForDir inspects dir and its parents to determine the
// version control system and code repository to use.
// On return, root is the import path
// corresponding to the root of the repository
// (thus root is a prefix of importPath).
func vcsForDir(dir, gopath string) (vcs *vcsCmd, root string, err error) {
	// Clean and double-check that dir is in (a subdirectory of) srcRoot.
	dir = filepath.Clean(dir)
	srcRoot := filepath.Clean(gopath)
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

	return nil, "", fmt.Errorf("directory %q is not using a known version control system", origDir)
}

// GetCommit provides the latest commit hash for a given directory.
// If a gopath is provided it limits the search inside the path.
func GetCommit(dir, gopath string) (string, error) {
	dir = filepath.Join(gopath, dir)
	vcs, root, err := vcsForDir(dir, gopath)
	if err != nil {
		return "", err
	}
	if vcs.commit == "" {
		return "", fmt.Errorf("%s is not yet supported", vcs.name)
	}
	args := strings.Split(vcs.commit, " ")
	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\n"), nil
}

// Checkout resets the head to hash commit for the given directory
func Checkout(dir, gopath, hash string) error {
	dir = filepath.Join(gopath, dir)
	vcs, root, err := vcsForDir(dir, gopath)
	if err != nil {
		return err
	}
	if vcs.checkout == "" {
		return fmt.Errorf("%s is not yet supported", vcs.name)
	}
	args := strings.Split(vcs.checkout, " ")
	args = append(args, hash)
	if len(args) == 1 {
		return Execute(root, args[0])
	}
	return Execute(root, args[0], args[1:]...)
}

// Fetch fetches all branches from remote
func Fetch(dir, root string) error {
	dir = filepath.Join(root, dir)
	vcs, root, err := vcsForDir(dir, root)
	if err != nil {
		return err
	}
	if vcs.fetch == "" {
		return fmt.Errorf("%s is not yet supported", vcs.name)
	}
	args := strings.Split(vcs.fetch, " ")
	if len(args) == 1 {
		return Execute(root, args[0])
	}
	return Execute(root, args[0], args[1:]...)
}

// Execute executes a command in the provided working directory
// and returns the stderr as error.
func Execute(cwd, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	errpipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	errout, err := ioutil.ReadAll(errpipe)
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("%s", errout)
	}
	return nil
}
