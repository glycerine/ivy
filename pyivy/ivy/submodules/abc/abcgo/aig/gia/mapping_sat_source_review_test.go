package gia

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"testing"
)

func TestMappingNativeSATOriginalFlow(t *testing.T) {
	exe := compileGiaSourceOracle(t, "if_sat_review.c")
	output, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("original SAT flow: %v %s", err, output)
	}
	want := "build XY 16 48\nand-result 1 assumptions 17 19 21 23 25 27 29 30\nempty-partition 1\none-var-set0 1\none-var-set1 1\none-var-set3 1\nsingle-output 1 123456789abcdef0\nunsat-outputs 0 1234 5678\nbuild XYZ 28 128\n"
	if string(output) != want {
		t.Fatalf("original SAT contract changed: %q want%q", output, want)
	}
}
