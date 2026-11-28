package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func randomMutationReviewGraph(mode int, signs, registers int32) *Gia_Man_t {
	g := Gia_ManStart(1)
	g.SetIdentity("mutation", "mutation.aig")
	switch mode {
	case 0:
		a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a^(signs&1), b^((signs>>1)&1))
		Gia_ManAppendCo(g, x)
		y := Gia_ManAppendAnd(g, c^((signs>>2)&1), d)
		z := Gia_ManAppendAnd(g, x^((signs>>3)&1), y)
		w := Gia_ManAppendAnd(g, 0, a)
		q := Gia_ManAppendAnd(g, z^((signs>>4)&1), w^1)
		duplicate := Gia_ManAppendAnd(g, a^(signs&1), b^((signs>>1)&1))
		r := Gia_ManAppendAnd(g, x, a^1)
		for _, lit := range []int32{q, z, w, a ^ 1, duplicate, r} {
			Gia_ManAppendCo(g, lit)
		}
		Gia_ManSetRegNum(g, registers)
	case 1, 2:
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		if mode == 1 {
			Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b^1))
		} else {
			Gia_ManAppendCo(g, Gia_ManAppendAnd(g, 0, a))
			Gia_ManAppendCo(g, b)
		}
	case 3:
		Gia_ManAppendCo(g, 0)
		Gia_ManAppendCo(g, 1)
	}
	return g
}

func randomMutationReviewRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	check := func(mode int, signs, regs int32) {
		t.Helper()
		src := randomMutationReviewGraph(mode, signs, regs)
		// Changing the caller's RNG state must not change this reset-based experiment.
		Gia_ManRandom(0)
		Gia_ManRandomW(0)
		out := Gia_ManDupPiPerm(src)
		next := Gia_ManRandom(0)
		wantNext := uint32(3176570839) // Reset, fourteen AND-edge draws, then draw sixteen.
		if mode == 1 || mode == 2 {
			wantNext = 3748422435
		} else if mode >= 3 {
			wantNext = 3633709067
		}
		if next != wantNext {
			t.Fatalf("mode %d: next random word %d, want %d", mode, next, wantNext)
		}
		if Gia_ManCiNum(out) != Gia_ManCiNum(src) || Gia_ManCoNum(out) != Gia_ManCoNum(src) || out.nRegs != src.nRegs || Gia_ManHashStarted(out) != 0 || Gia_ManHasDangling(out) != 0 {
			t.Fatal("random mutation changed interface or retained transient state")
		}
		if name, spec := out.Identity(); name != "mutation" || spec != "mutation.aig" {
			t.Fatal("random mutation lost source identity")
		}
		fmt.Fprintf(&records, "%d %d %d %d %d %d %d %d", mode, signs, regs, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs, next)
		for mask := 0; mask < 1<<uint(Gia_ManCiNum(out)); mask++ {
			inputs := make([]bool, Gia_ManCiNum(out))
			for i := range inputs {
				inputs[i] = mask&(1<<uint(i)) != 0
			}
			got, err := out.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			if mode == 0 {
				a, c, d := inputs[0], inputs[2], inputs[3]
				x := (d != (signs&1 != 0)) && (c != (signs&2 != 0))
				y := (d != (signs&4 != 0)) && c
				z := (x != (signs&8 != 0)) && y
				w := c && d
				qInput := z != (signs&16 != 0)
				// Equal hashed x/y with an inverted x fold z to literal zero.
				// Source mutation then replaces that literal on q's edge too.
				if signs&2 == 0 && (signs&1) == ((signs>>2)&1) && signs&8 != 0 {
					qInput = c != (signs&16 != 0)
				}
				checkQuantificationOutputs(t, out, inputs, []bool{x, qInput && !w, z, w, !a, x, !a && x})
			}
			packed := 0
			for i, value := range got {
				if value {
					packed |= 1 << uint(i)
				}
			}
			fmt.Fprintf(&records, " %d", packed)
		}
		records.WriteByte('\n')
	}
	for signs := int32(0); signs < 32; signs++ {
		for _, regs := range []int32{0, 2, 4} {
			check(0, signs, regs)
		}
	}
	for mode := 1; mode <= 4; mode++ {
		check(mode, 0, 0)
	}
	return records.String()
}

func TestRandomMutationResetAndTruthTables(t *testing.T) {
	randomMutationReviewRecords(t)
}

func TestRandomMutationMatchesOriginalSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "random_mutation.c")
	want, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C random mutation: %v: %s", err, want)
	}
	gotRows := strings.Split(strings.TrimSpace(randomMutationReviewRecords(t)), "\n")
	wantRows := strings.Split(strings.TrimSpace(string(want)), "\n")
	if len(gotRows) != len(wantRows) {
		t.Fatalf("Go cases %d, C cases %d", len(gotRows), len(wantRows))
	}
	for i := range gotRows {
		if gotRows[i] != wantRows[i] {
			t.Fatalf("case %d:\nGo %s\nC  %s", i, gotRows[i], wantRows[i])
		}
	}
}

func TestRandomMutationSourceGoldenCases(t *testing.T) {
	for mode := 1; mode <= 4; mode++ {
		src := randomMutationReviewGraph(mode, 0, 0)
		out := Gia_ManDupPiPerm(src)
		wantNext := uint32(3748422435) // Reset draw + two AND-edge draws; next is draw four.
		if mode >= 3 {
			wantNext = 3633709067 // No internal gates: only the reset draw was consumed.
		}
		if got := Gia_ManRandom(0); got != wantNext {
			t.Fatalf("mode %d: next random word %d, want %d", mode, got, wantNext)
		}
		for mask := 0; mask < 1<<uint(Gia_ManCiNum(out)); mask++ {
			var inputs, want []bool
			if mode <= 2 {
				a, b := mask&1 != 0, mask&2 != 0
				inputs = []bool{a, b}
				want = []bool{!a && b}
				if mode == 2 {
					// The source's <= CI-count condition also replaces constant zero.
					want = []bool{a && b, b}
				}
			} else if mode == 3 {
				want = []bool{false, true}
			}
			checkQuantificationOutputs(t, out, inputs, want)
		}
	}
}

func TestRandomMutationTypedLogicalPins(t *testing.T) {
	// Original C's AND-only experiment has no native real-gate contract.
	// Check the existing Go extension's logical-pin draws independently.
	for signs := int32(0); signs < 8; signs++ {
		src := quantificationReviewGraph(signs, false, true)
		out := Gia_ManDupPiPerm(src)
		if next := Gia_ManRandom(0); next != 2777078678 { // Draw ten after eight logical-pin draws.
			t.Fatalf("typed mutation consumed the wrong random stream: %d", next)
		}
		for mask := 0; mask < 16; mask++ {
			a, b, c, d := mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0
			x := (d != c) != (signs&1 != 0)
			buffer := x != (signs&2 != 0)
			mux := c
			if c != (signs&4 != 0) {
				mux = buffer
			}
			checkQuantificationOutputs(t, out, []bool{a, b, c, d}, []bool{!d && mux})
		}
	}
}

func TestRandomMutationWithoutCisCopiesCollapsedLiterals(t *testing.T) {
	// The C experiment would divide by zero on internal gates with no CIs.
	// Retain the existing Go fallback, including literals folded by earlier gates.
	src := Gia_ManStart(1)
	src.fGiaSimple = 1
	buffer := Gia_ManAppendBuf(src, 1)
	zero := Gia_ManAppendAnd(src, buffer, buffer^1)
	Gia_ManAppendCo(src, buffer)
	Gia_ManAppendCo(src, zero)
	out := Gia_ManDupPiPerm(src)
	if next := Gia_ManRandom(0); next != 3633709067 {
		t.Fatalf("no-CI fallback consumed unexpected draws: %d", next)
	}
	if out.nObjs != 3 || Gia_ManCiNum(out) != 0 || out.nRegs != 0 {
		t.Fatal("no-CI fallback retained unnecessary logic")
	}
	checkQuantificationOutputs(t, out, nil, []bool{true, false})
}
