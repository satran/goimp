package goimp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
	vcs, root, err := vcsForDir(dir, gopath)
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
