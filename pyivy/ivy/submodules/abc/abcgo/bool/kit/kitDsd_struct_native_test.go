package kit

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"
)

// test_vectors/dsd_struct_native.txt is printed by dsd_struct_native.c
// linked against native ABC: random decomposable functions and the DSD
// networks Kit_DsdDecompose, Kit_DsdDecomposeExpand and the expanded
// Kit_DsdDecomposeMux build for them. lutpack bases its bound sets and
// LUT functions on these networks, so their structure and literal phases
// must match ABC's.
func TestDsdDecomposeMatchesNative(t *testing.T) {
	f, err := os.Open("test_vectors/dsd_struct_native.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	bad := 0
	for line := 1; scanner.Scan(); line++ {
		parts := strings.Split(scanner.Text(), " | ")
		fields := strings.Fields(parts[0])
		nVars, _ := strconv.Atoi(fields[0])
		truth := make([]uint32, 8)
		for w, hex := range fields[1:] {
			v, _ := strconv.ParseUint(hex, 16, 32)
			truth[w] = uint32(v)
		}
		got := make([]string, 3)
		for i, ntk := range []*Kit_DsdNtk_t{
			Kit_DsdDecompose(truth, int32(nVars)),
			Kit_DsdDecomposeExpand(truth, int32(nVars)),
			Kit_DsdExpand(Kit_DsdDecomposeMux(truth, int32(nVars), 8)),
		} {
			var b bytes.Buffer
			Kit_DsdPrint(&b, ntk)
			got[i] = b.String()
		}
		for i, name := range []string{"Decompose", "DecomposeExpand", "DecomposeMux+Expand"} {
			if got[i] != parts[1+i] {
				if bad++; bad <= 5 {
					t.Errorf("line %d %s(%s): got %s want %s", line, name, parts[0], got[i], parts[1+i])
				}
			}
		}
	}
	if bad != 0 {
		t.Fatalf("%d mismatches", bad)
	}
}
