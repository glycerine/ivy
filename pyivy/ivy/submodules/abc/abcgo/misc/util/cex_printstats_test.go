package util

import (
	"io"
	"os"
	"testing"
)

// TestCexPrintStatsLikeNative: Abc_CexPrintStats prints utilCex.c's line.
func TestCexPrintStatsLikeNative(t *testing.T) {
	c := Abc_CexAlloc(3, 4, 1)
	Abc_InfoSetBit(c.PData, 4)
	r, w, _ := os.Pipe()
	saved := os.Stdout
	os.Stdout = w
	Abc_CexPrintStats(c)
	Abc_CexPrintStats(nil)
	w.Close()
	os.Stdout = saved
	out, _ := io.ReadAll(r)
	want := "CEX: Po =   0  Frame =   0  FF = 3  PI = 4  Bit =       7  1s =       1 (25.00 %)\nThe counter example is NULL.\n"
	if string(out) != want {
		t.Fatalf("got %q\nwant %q", out, want)
	}
}
