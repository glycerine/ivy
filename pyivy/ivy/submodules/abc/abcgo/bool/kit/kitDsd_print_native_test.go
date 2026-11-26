package kit

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"
)

// test_vectors/dsd_print_native.txt is printed by dsd_print_native.c
// linked against native ABC: a truth table, then Kit_DsdPrintFromTruth's
// and Kit_DsdWriteFromTruth's text for it.
func TestDsdPrintMatchesNative(t *testing.T) {
	f, err := os.Open("test_vectors/dsd_print_native.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for line := 1; scanner.Scan(); line++ {
		parts := strings.Split(scanner.Text(), " | ")
		fields := strings.Fields(parts[0])
		nVars, _ := strconv.Atoi(fields[0])
		truth := make([]uint32, 8)
		for w, hex := range fields[1:] {
			v, _ := strconv.ParseUint(hex, 16, 32)
			truth[w] = uint32(v)
		}
		var printed bytes.Buffer
		Kit_DsdPrintFromTruth2(&printed, truth, int32(nVars))
		if printed.String() != parts[1] {
			t.Fatalf("line %d print: got %q want %q", line, printed.String(), parts[1])
		}
		if got := Kit_DsdWriteFromTruth(truth, int32(nVars)); got != parts[2] {
			t.Fatalf("line %d write: got %q want %q", line, got, parts[2])
		}
	}
}
