package ver

import (
	"os"
	"testing"
)

func TestSignalRangeSourceCases(t *testing.T) {
	p := &Ver_Man_t{Output: os.Stdout, fTopLevel: 1}
	for _, tc := range []struct {
		in, rest string
		hi, lo   int32
	}{{"[7:2]bus", "bus", 7, 2}, {"[7:2]\\escaped name", "escaped", 7, 2}, {"[3]bus", "]bus", 3, 3}, {"[2:7]x", "x", 2, 7}} {
		word := tc.in
		var hi, lo int32
		if Ver_ParseSignalPrefix(p, &word, &hi, &lo) != 1 || word != tc.rest || hi != tc.hi || lo != tc.lo {
			t.Fatalf("prefix %q: %q %d %d", tc.in, word, hi, lo)
		}
	}
	for _, tc := range []struct {
		in, rest string
		hi, lo   int32
	}{{"bus[7:2]", "bus", 7, 2}, {"x[3]", "x", 3, 3}, {"x[2:7]", "x", 2, 7}} {
		word := tc.in
		var hi, lo int32
		if Ver_ParseSignalSuffix(p, &word, &hi, &lo) != 1 || word != tc.rest || hi != tc.hi || lo != tc.lo {
			t.Fatalf("suffix %q: %q %d %d", tc.in, word, hi, lo)
		}
	}
}
