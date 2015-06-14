package main

import "testing"

var (
	testStdLibPaths = []string{"fmt", "go/ast", "errors"}
	testOtherPaths  = []string{
		"github.com/optiopay/kafka",
		"github.com/satran/edi",
	}
)

func TestIsStdLib(t *testing.T) {
	for _, path := range testStdLibPaths {
		if !isStdLib(path) {
			t.Fatalf("'%s' is a stdlib package", path)
		}
	}
	for _, path := range testOtherPaths {
		if isStdLib(path) {
			t.Fatalf("'%s' is not stdlib package", path)
		}
	}
}
