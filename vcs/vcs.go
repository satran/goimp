package vcs

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type VCS struct {
	// Root is the import path corresponding to the root of the repository
	Root string

	name     string
	cmd      string
	commit   string
	checkout string
	fetch    string
	pull     string
}

var vcsList = []VCS{
	{
		name:     "Git",
		cmd:      "git",
		commit:   "rev-parse HEAD",
		checkout: "checkout",
		fetch:    "fetch",
		pull:     "pull",
	},
	{
		name:     "Mercurial",
		cmd:      "hg",
		commit:   "id -i",
		checkout: "update",
		fetch:    "pull",
		pull:     "pull -u",
	},
	{
		name:     "Bazaar",
		cmd:      "bzr",
		commit:   "revno",
		checkout: "revert -r",
		fetch:    "pull --overwrite",
	},
}

// New inspects dir and its parents to determine the
// version control system and code repository to use.
func New(dir, parent string) (vcs *VCS, err error) {
	// Clean and double-check that dir is in (a subdirectory of) srcRoot.
	dir = filepath.Clean(dir)
	srcRoot := filepath.Clean(parent)
	if len(dir) <= len(srcRoot) || dir[len(srcRoot)] != filepath.Separator {
		return nil, fmt.Errorf(
			"directory %q is outside source root %q", dir, srcRoot)
	}

	origDir := dir
	for len(dir) > len(srcRoot) {
		for _, vcs := range vcsList {
			fi, err := os.Stat(filepath.Join(dir, "."+vcs.cmd))
			if err == nil && fi.IsDir() {
				v := vcs
				v.Root = dir
				return &v, nil
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

	return nil, fmt.Errorf("directory %q is not using a known version control system", origDir)
}

// CommitHash provides the latest commit hash for a given directory.
// If a gopath is provided it limits the search inside the path.
func (v *VCS) CommitHash() (string, error) {
	if v.commit == "" {
		return "", fmt.Errorf("%s is not yet supported", v.name)
	}
	args := strings.Split(v.commit, " ")
	cmd := exec.Command(v.cmd, args...)
	cmd.Dir = v.Root
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\n"), nil
}

// Checkout resets the head to hash commit for the given directory
func (v *VCS) Checkout(hash string) error {
	if v.checkout == "" {
		return fmt.Errorf("%s is not yet supported", v.name)
	}
	current, err := v.CommitHash()
	if err != nil {
		return err
	}
	if current == hash {
		return nil
	}
	args := strings.Split(v.checkout, " ")
	args = append(args, hash)
	return execute(v.Root, v.cmd, args...)
}

// Fetch fetches all branches from remote
func (v *VCS) Fetch() error {
	if v.fetch == "" {
		return fmt.Errorf("%s is not yet supported", v.name)
	}
	args := strings.Split(v.fetch, " ")
	return execute(v.Root, v.cmd, args...)
}

// Latest gets the latest code from remote
func (v *VCS) Latest() error {
	if v.pull == "" {
		return fmt.Errorf("%s is not yet supported", v.name)
	}
	if v.cmd == "git" {
		err := execute(v.Root, "git", "checkout", "master")
		if err != nil {
			return err
		}
	}
	args := strings.Split(v.pull, " ")
	return execute(v.Root, v.cmd, args...)
}

// execute executes a command in the provided working directory
// and returns the stderr as error.
func execute(cwd, command string, args ...string) error {
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
