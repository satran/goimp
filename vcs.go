package goimp

import (
	"fmt"
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
}

var vcsList = []*vcsCmd{
	{
		name:     "Git",
		cmd:      "git",
		commit:   "git rev-parse HEAD",
		checkout: "git checkout",
	},
	{
		name:     "Mercurial",
		cmd:      "hg",
		commit:   "hg id -i",
		checkout: "hg update",
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
	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", output, err)
	}
	return nil
}
