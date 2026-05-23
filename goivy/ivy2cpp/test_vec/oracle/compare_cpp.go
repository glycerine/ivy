package main

import (
	"fmt"
	"os"

	"github.com/glycerine/ivy/goivy/ivy2cpp"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: compare_cpp left.cpp right.cpp\n")
		os.Exit(2)
	}
	left, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", os.Args[1], err)
		os.Exit(2)
	}
	right, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", os.Args[2], err)
		os.Exit(2)
	}
	diff := ivy2cpp.CompareCPPTokens(os.Args[1], string(left), os.Args[2], string(right))
	if !diff.Equal {
		fmt.Fprintln(os.Stderr, diff.Error())
		os.Exit(1)
	}
}
