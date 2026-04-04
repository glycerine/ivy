//go:build !xtracer_off

package xtracer

import (
	"fmt"
	"testing"
)

func TestNormalizeLine_IvyExamples(t *testing.T) {
	line := "/Users/jaten/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy"
	line2 := NormalizeLine(line)
	if line2 == line {
		//fmt.Printf("bad! same: %v\n", line2)
		t.Fatalf("IVY_EXAMPLES replacement failed!")
	} else {
		fmt.Printf("good: %v\n", line2)
	}
}
