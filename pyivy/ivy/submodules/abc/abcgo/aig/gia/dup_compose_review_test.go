package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func cascadeReviewSecond(signs int32) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	x := Gia_ManAppendXorReal(g, a^(signs&1), b)
	buffer := Gia_ManAppendBuf(g, x^((signs>>1)&1))
	m := Gia_ManAppendMuxReal(g, c^((signs>>2)&1), buffer, a^1)
	Gia_ManAppendCo(g, m)
	Gia_ManAppendCo(g, x^1)
	Gia_ManAppendAnd(g, b, c) // Deliberately dangling; cleanup must remove it.
	return g
}

func cascadeReviewRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for signs1 := int32(0); signs1 < 8; signs1++ {
		for signs2 := int32(0); signs2 < 8; signs2++ {
			first := quantificationReviewGraph(signs1, false, false)
			second := cascadeReviewSecond(signs2)
			out := Gia_ManDupOntop(first, second)
			if Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != 2 || out.nRegs != 0 {
				t.Fatal("cascade changed interface")
			}
			if Gia_ManHasDangling(out) != 0 {
				t.Fatal("cascade retained dangling logic")
			}
			if name, spec := out.Identity(); name != "quantification" || spec != "quantification.aig" {
				t.Fatal("cascade lost first-stage identity")
			}
			fmt.Fprintf(&records, "%d %d %d %d %d", signs1, signs2, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				intermediate, err := first.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				want, err := second.Evaluate(intermediate)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, out, inputs, want)
				values, err := out.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				packed := 0
				for i, value := range values {
					if value {
						packed |= 1 << uint(i)
					}
				}
				fmt.Fprintf(&records, " %d", packed)
			}
			records.WriteByte('\n')
		}
	}
	return records.String()
}

func TestCascadeTypedTruth(t *testing.T) {
	cascadeReviewRecords(t)
}

func TestCascadeMatchesOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cascade.c")
	wantC, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C cascade: %v: %s", err, wantC)
	}
	gotRows := strings.Split(strings.TrimSpace(cascadeReviewRecords(t)), "\n")
	wantRows := strings.Split(strings.TrimSpace(string(wantC)), "\n")
	if len(gotRows) != len(wantRows) {
		t.Fatalf("Go cases %d, C cases %d", len(gotRows), len(wantRows))
	}
	for i := range gotRows {
		if gotRows[i] != wantRows[i] {
			t.Fatalf("case %d:\nGo %s\nC  %s", i, gotRows[i], wantRows[i])
		}
	}
}

func TestCascadeAcceptsSameGraphForBothStages(t *testing.T) {
	for signs := int32(0); signs < 4; signs++ {
		g := Gia_ManStart(1)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		Gia_ManAppendCo(g, a^(signs&1))
		Gia_ManAppendCo(g, Gia_ManAppendXorReal(g, a, b^((signs>>1)&1)))
		out := Gia_ManDupOntop(g, g)
		for mask := 0; mask < 4; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0}
			intermediate, err := g.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			want, err := g.Evaluate(intermediate)
			if err != nil {
				t.Fatal(err)
			}
			checkQuantificationOutputs(t, out, inputs, want)
		}
	}
}
