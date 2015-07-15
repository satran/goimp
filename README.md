# goimp

Yet another dependency manager for golang.

### Install
Use the INSTALL script provided in the repo. If you don't want support
for vendor images just run 'go install' or
'go get github.com/satran/goimp'.

The `--help` option in goimp provides documentation about the various
sub-commands and options.

### Vendor Images
Goimp provides a simple script that can be sourced to create a
environment for your project. To use it run

    source goimp-init

This will create a vendor.img in the package root. This will be
mounted on a temporary directory and the `GOROOT` variable set to the
temporary directory. 