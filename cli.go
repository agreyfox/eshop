// Package main is located in the cmd/ponzu directory and contains the code to build
// and operate the command line interface (CLI) to manage Ponzu systems. Here,
// you will find the code that is used to create new Ponzu projects, generate
// code for content types and other files, build Ponzu binaries and run servers.
package main

import (
	"runtime"

	"github.com/agreyfox/eshop/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}
