// Command abc-starter executes a file of shell commands in parallel.
package main

import (
	"github.com/glycerine/abcgo"
	"os"
)

func main() { os.Exit(abcgo.StarterMain(os.Args[1:], os.Stdout, os.Stderr)) }
