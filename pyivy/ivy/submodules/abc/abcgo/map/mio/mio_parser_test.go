package mio

import (
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
	"testing"
)

func TestFormulaPrecedenceConstantsAndPolarity(t *testing.T) {
	names := []string{"a", "b", "c"}
	cases := []struct {
		formula string
		eval    func(bool, bool, bool) bool
	}{
		{"a + b*c", func(a, b, c bool) bool { return a || b && c }},
		{"a ^ b + c", func(a, b, c bool) bool { return a != b || c }},
		{"!(a b') ^ c", func(a, b, c bool) bool { return !(a && !b) != c }},
		{"(a+CONST0)*1", func(a, b, c bool) bool { return a }},
		{"!!a''", func(a, b, c bool) bool { return a }},
	}
	for _, test := range cases {
		p, err := ParseFormula(test.formula, names)
		if err != nil {
			t.Fatal(test.formula, err)
		}
		truth := Exp_Truth6(3, p, nil)
		sop := Mio_LibDeriveSop(3, p, nil)
		for minterm := 0; minterm < 64; minterm++ {
			want := test.eval(minterm&1 != 0, minterm&2 != 0, minterm&4 != 0)
			got := truth>>minterm&1 != 0
			if got != want {
				t.Fatal(test.formula, minterm, truth)
			}
			if evaluateTestSop(sop, minterm) != want {
				t.Fatal("SOP mismatch", test.formula, sop, minterm)
			}
		}
	}
	for _, formula := range []string{"", "a+", "(a", "a)", "missing", "a**b", "a;", "'a"} {
		if _, err := ParseFormula(formula, names); err == nil {
			t.Fatal("accepted malformed formula", formula)
		}
	}
}
func evaluateTestSop(sop string, minterm int) bool {
	polarity := true
	matched := false
	for _, line := range strings.Split(strings.TrimSuffix(sop, "\n"), "\n") {
		index := strings.LastIndexByte(line, ' ')
		if index < 0 {
			panic("bad test SOP")
		}
		polarity = line[index+1] == '1'
		cube := line[:index]
		hit := true
		for i, c := range cube {
			if c == '0' && minterm>>i&1 != 0 || c == '1' && minterm>>i&1 == 0 {
				hit = false
			}
		}
		matched = matched || hit
	}
	if polarity {
		return matched
	}
	return !matched
}
func TestExpressionWideTruthAndSOPSubsumption(t *testing.T) {
	names := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	p, err := ParseFormula("a ^ h", names)
	if err != nil {
		t.Fatal(err)
	}
	truth := make([]uint64, 4)
	Exp_Truth8(8, p, nil, truth)
	generic := make([]uint64, 4)
	Exp_Truth(8, p, generic)
	for m := 0; m < 256; m++ {
		want := (m&1 != 0) != (m&128 != 0)
		if (truth[m/64]>>uint(m%64)&1 != 0) != want || truth[m/64] != generic[m/64] {
			t.Fatal(m, truth, generic)
		}
	}
	cover := Mio_SopConst0()
	Mio_SopPushSCC(cover, Mio_CubeVar0(0)|Mio_CubeVar1(1))
	Mio_SopPushSCC(cover, Mio_CubeVar0(0))
	if vec.Vec_IntSize(cover) != 1 {
		t.Fatal("subsumption")
	}
	if vec.Vec_IntSize(Mio_SopCoverAnd(Mio_SopVar0(0), Mio_SopVar1(0))) != 0 {
		t.Fatal("contradictory cube")
	}
}

const testGenlib = `# real genlib records
GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 2 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 2 0
GATE and2 2 O=a*b;
 PIN b NONINV 2 999 3 0 4 0
 PIN a NONINV 1 999 1 0 2 0
GATE xor2 3 O=a^b; PIN * UNKNOWN 1 999 2 0 2 0
.end
`

func TestGenlibDerivedDataAndErrors(t *testing.T) {
	lib, err := ReadLibrary("fixture", testGenlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	if lib.nGates != 6 || lib.pGate0 == nil || lib.pGate1 == nil || lib.pGateInv == nil || lib.pGateBuf == nil || lib.pGateAnd2 == nil {
		t.Fatal("special gates", lib)
	}
	and := Mio_LibraryReadGateByName(lib, "and2", "")
	if and.pPins.pName != "b" || and.pPins.pNext.pName != "a" || and.dDelayMax != 4 || and.nInputs != 2 || and.uTruth != 0x8888888888888888 {
		t.Fatal("derived", and)
	}
	if lib.ppGates0[0].pName != "zero" || lib.ppGatesName[0].pName != "and2" {
		t.Fatal("sort and original IDs")
	}
	filtered, err := ReadLibrary("exclude", testGenlib, map[string]bool{"xor2": true})
	if err != nil || filtered.nGates != 5 {
		t.Fatal("exclusion", err)
	}
	bad := []string{"", "GATE x NaN O=a; PIN * INV 1 1 1 1 1 1", "GATE x 1 O=a PIN * INV 1 1 1 1 1 1", "GATE x 1 O=a; PIN z INV 1 1 1 1 1 1", "GATE x 1 O=a; PIN * BAD 1 1 1 1 1 1", "GATE x 1 O=a; PIN * INV 1 1 1", "GATE x 1 O=a; PIN a INV 1 1 1 1 1 1 PIN a INV 1 1 1 1 1 1", "GATE x 1 O=0; .end garbage"}
	for _, text := range bad {
		if _, err := ReadLibrary("bad", text, nil); err == nil {
			t.Fatal("accepted malformed genlib", text)
		}
	}
}
func TestGenlibExtendedAndTwinOutputs(t *testing.T) {
	lib, err := ReadLibrary("extended", "GATE x 1 O=a; PIN * INV 1 2 3 4 99 5 6 99", nil)
	if err != nil {
		t.Fatal(err)
	}
	pin := lib.pGates.pPins
	if pin.dDelayBlockFall != 5 || pin.dDelayFanoutFall != 6 {
		t.Fatal(pin)
	}
	twin := `GATE pair 1 X=a; PIN * NONINV 1 1 1 1 1 1
GATE pair 1 Y=!a; PIN * INV 1 1 1 1 1 1`
	lib, err = ReadLibrary("twins", twin, nil)
	if err != nil {
		t.Fatal(err)
	}
	g := Mio_LibraryReadGateByName(lib, "pair", "Y")
	if g == nil || g.pTwin == nil || g.pOutName != "Y" {
		t.Fatal("twin lookup")
	}
	if _, err = ReadLibrary("third", twin+"\nGATE pair 1 Z=a; PIN * NONINV 1 1 1 1 1 1", nil); err == nil {
		t.Fatal("accepted third output")
	}
}
