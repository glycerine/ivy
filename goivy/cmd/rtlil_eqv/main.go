package main

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/glycerine/ivy/goivy/rtlileqv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: rtlil_eqv <config>.ileq")
		os.Exit(2)
	}
	start := time.Now()
	result, err := rtlileqv.Check(context.Background(), os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	cones := result.Outputs + result.Registers
	if len(result.Failures) == 0 {
		fmt.Printf("compared %d cones (%d outputs, %d registers) in %.2fs\nOK\n", cones, result.Outputs, result.Registers, time.Since(start).Seconds())
		return
	}
	fmt.Printf("NOT EQUIVALENT: %d of %d cones differ\n\n", len(result.Failures), cones)
	for _, failure := range result.Failures[:min(20, len(result.Failures))] {
		fmt.Printf("cone: %s\n", failure.Cone)
		for _, name := range slices.Sorted(maps.Keys(failure.Inputs)) {
			bit := 0
			if failure.Inputs[name] {
				bit = 1
			}
			fmt.Printf("    %-28s = %d\n", name, bit)
		}
		fmt.Println()
	}
	if extra := len(result.Failures) - 20; extra > 0 {
		fmt.Printf("... and %d more differing cones\n", extra)
	}
	os.Exit(1)
}
