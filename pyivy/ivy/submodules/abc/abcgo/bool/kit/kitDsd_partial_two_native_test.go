package kit

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// test_vectors/dsd_partial_two_native.txt is the first 3000 lines printed by
// test_vectors/dsd_partial_two_native.c linked against native ABC: random
// functions, a bound set, and Kit_DsdTruthPartialTwo's composition and
// decomposition truth tables. Lpk_DsdSplit builds LUTs from both, so their
// polarity and the vacuous-variable placement must match ABC exactly.
func TestDsdTruthPartialTwoMatchesNative(t *testing.T) {
	f, err := os.Open("test_vectors/dsd_partial_two_native.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	hex := func(s string) uint32 {
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			t.Fatal(err)
		}
		return uint32(v)
	}
	m := Kit_DsdManAlloc(8, 64)
	// Like the C driver, reuse the output buffers: some paths leave the
	// decomposition function untouched.
	co, dec := make([]uint32, 8), make([]uint32, 8)
	scanner := bufio.NewScanner(f)
	cases := 0
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " | ")
		head := strings.Fields(parts[0])
		nVars, _ := strconv.Atoi(head[1])
		bound := hex(head[2])
		words := int(Kit_TruthWordNum(int32(nVars)))
		truth, wantCo, wantDec := make([]uint32, 8), strings.Fields(parts[1]), strings.Fields(parts[2])
		for w := 0; w < words; w++ {
			truth[w] = hex(head[3+w])
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("case %s: panic %v", head[0], r)
				}
			}()
			ntk := Kit_DsdDecomposeExpand(truth, int32(nVars))
			Kit_DsdTruthPartialTwo(m, ntk, bound, Kit_WordFindFirstBit(bound), co, dec)
		}()
		for w := 0; w < words; w++ {
			if co[w] != hex(wantCo[w]) || dec[w] != hex(wantDec[w]) {
				t.Fatalf("case %s (vars %d, bound %x): got co %08x dec %08x, want %s %s", head[0], nVars, bound, co[:words], dec[:words], wantCo, wantDec)
			}
		}
		cases++
	}
	if cases != 3000 {
		t.Fatal("cases", cases)
	}
}
