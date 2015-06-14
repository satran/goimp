#goimp

A very simple dependency manager for golang.

```
Usage goimp:
  -debug=false: error logs are prefixed with file name
  -dir=".": project source directory
  -hash=false: adds commit hash when writing or listing
  -l=false: lists dependency to stdout
  -r=false: finds imports recursively
  -w=false: writes dependency to deps file
```

It uses the first part of your `GOPATH` env variable to set the source
files to the commit hash that is required by your code.

`goimp` without any arguments looks for a `deps` file. It fetches the
latest commit and checks out to the commit if provided in the deps
file.
