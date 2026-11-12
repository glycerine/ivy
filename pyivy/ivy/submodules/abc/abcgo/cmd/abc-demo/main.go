// Command abc-demo runs ABC's library demonstration on one input network.
package main

import (
	"github.com/glycerine/abcgo"
	"os"
)

func main() { os.Exit(abcgo.DemoMain(os.Args[1:], os.Stdout, os.Stderr)) }
