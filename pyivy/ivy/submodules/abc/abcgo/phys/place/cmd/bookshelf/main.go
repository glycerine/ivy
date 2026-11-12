// Bookshelf runs ABC's original three-file global-placement harness.
package main

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/phys/place"
)

func main() {
	if err := place.RunBookshelf(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
