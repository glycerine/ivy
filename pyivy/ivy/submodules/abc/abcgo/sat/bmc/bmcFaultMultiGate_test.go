package bmc

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// This oracle deliberately evaluates gates directly, without using the fault
// unfolding or GIA simulation implementations under test.
func TestFaultDriverTwoGateFaultCoverage(t *testing.T) {
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, x, c))
	for _, tc := range []struct {
		name      string
		algo      int32
		form      string
		basic     int32
		ncontrols int
	}{
		{"formula flip", 0, "((a&b)^p)", 0, 2},
		{"flip single", 3, "", 1, 2},
		{"flip multiple", 3, "", 0, 2},
		{"stuck single", 2, "", 1, 4},
		{"stuck multiple", 2, "", 0, 4},
		{"function single", 4, "", 1, 8},
		{"function multiple", 4, "", 0, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			os.Remove("tests.txt")
			Gia_ManFaultTest(g, g, &Bmc_ParFf_t{Algo: tc.algo, PFormStr: tc.form, FBasic: tc.basic, FDump: 1})
			data, err := os.ReadFile("tests.txt")
			if err != nil {
				t.Fatal(err)
			}
			var suite []int
			for _, row := range strings.Fields(string(data)) {
				if len(row) != 3 {
					t.Fatalf("invalid pattern %q", row)
				}
				bits := 0
				for i, ch := range row {
					if ch == '1' {
						bits |= 1 << i
					} else if ch != '0' {
						t.Fatalf("invalid pattern %q", row)
					}
				}
				suite = append(suite, bits)
			}
			evaluate := func(in, controls int) bool {
				first := (in&1 != 0) && (in&2 != 0)
				if tc.algo == 4 {
					row := 0
					if in&1 != 0 {
						row++
					}
					if in&2 != 0 {
						row += 2
					}
					first = first != (controls&(1<<row) != 0)
				} else if tc.algo == 2 {
					first = first && controls&1 == 0
					first = first || controls&2 != 0
				} else {
					first = first != (controls&1 != 0)
				}
				second := first && (in&4 != 0)
				if tc.algo == 4 {
					row := 0
					if in&4 != 0 {
						row++
					}
					if first {
						row += 2
					}
					if row != 3 {
						second = second != (controls&(1<<(4+row)) != 0)
					}
				} else if tc.algo == 2 {
					second = second && controls&4 == 0
					second = second || controls&8 != 0
				} else {
					second = second != (controls&2 != 0)
				}
				return second
			}
			for mask := 1; mask < 1<<tc.ncontrols; mask++ {
				if tc.basic != 0 && mask&(mask-1) != 0 {
					continue
				}
				detectable, detected := false, false
				for in := 0; in < 8; in++ {
					if evaluate(in, 0) != evaluate(in, mask) {
						detectable = true
					}
				}
				for _, in := range suite {
					if evaluate(in, 0) != evaluate(in, mask) {
						detected = true
					}
				}
				if detectable && !detected {
					t.Errorf("suite %q misses fault mask %0*b", data, tc.ncontrols, mask)
				}
			}
		})
	}
}

func TestFaultDriverTwoGateDelayCoverage(t *testing.T) {
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	g := gia.Gia_ManStart(16)
	a, b, r := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, r)
	gia.Gia_ManAppendCo(g, y)
	gia.Gia_ManAppendCo(g, y)
	gia.Gia_ManSetRegNum(g, 1)
	for _, ffOnly := range []int32{0, 1} {
		for _, basic := range []int32{0, 1} {
			t.Run(fmt.Sprintf("ff%d_basic%d", ffOnly, basic), func(t *testing.T) {
				os.Remove("tests.txt")
				Gia_ManFaultTest(g, g, &Bmc_ParFf_t{Algo: 1, FFfOnly: ffOnly, FBasic: basic, FDump: 1})
				data, err := os.ReadFile("tests.txt")
				if err != nil {
					t.Fatal(err)
				}
				var suite []int
				for _, row := range strings.Fields(string(data)) {
					if len(row) != 5 {
						t.Fatalf("invalid pattern %q", row)
					}
					bits := 0
					for i, ch := range row {
						if ch == '1' {
							bits |= 1 << i
						} else if ch != '0' {
							t.Fatalf("invalid pattern %q", row)
						}
					}
					suite = append(suite, bits)
				}
				eval := func(bits, mask int) (bool, bool) {
					r0, a0, b0, a1, b1 := bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0, bits&16 != 0
					x0 := a0 && b0
					y0 := x0 && r0
					x1 := a1 && b1
					if ffOnly == 0 && mask&1 != 0 {
						x1 = x0
					}
					y1 := x1 && y0
					if ffOnly == 0 && mask&2 != 0 {
						y1 = y0
					}
					po, ri := y1, y1
					if ffOnly != 0 && mask&1 != 0 {
						ri = y0
					}
					return po, ri
				}
				n := 2
				if ffOnly != 0 {
					n = 1
				}
				for mask := 1; mask < 1<<n; mask++ {
					if basic != 0 && mask&(mask-1) != 0 {
						continue
					}
					detectable, detected := false, false
					for bits := 0; bits < 32; bits++ {
						a, b := eval(bits, 0)
						c, d := eval(bits, mask)
						if a != c || b != d {
							detectable = true
						}
					}
					for _, bits := range suite {
						a, b := eval(bits, 0)
						c, d := eval(bits, mask)
						if a != c || b != d {
							detected = true
						}
					}
					if detectable && !detected {
						t.Errorf("suite %q misses mask %b", data, mask)
					}
				}
			})
		}
	}
}
