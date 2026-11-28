package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func appendReviewGraph(signs, registers int32) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ros := make([]int32, registers)
	for i := range ros {
		ros[i] = Gia_ManAppendCi(g)
	}
	control, other := a, b
	if registers != 0 {
		control = ros[0]
	}
	if registers > 1 {
		other = ros[len(ros)-1]
	}
	g.EnableMuxes()
	x := Gia_ManAppendXorReal(g, a^(signs&1), b)
	buffer := Gia_ManAppendBuf(g, x^((signs>>1)&1))
	mux := Gia_ManAppendMuxReal(g, control^((signs>>2)&1), buffer, other)
	Gia_ManAppendCo(g, mux)
	Gia_ManAppendCo(g, a^1)
	for _, ro := range ros {
		Gia_ManAppendCo(g, Gia_ManAppendXorReal(g, ro, mux))
	}
	Gia_ManSetRegNum(g, registers)
	g.SetIdentity("append", "append.aig")
	return g
}

func checkAppendReviewRow(t *testing.T, out, first, second *Gia_Man_t, inputs []bool) {
	t.Helper()
	boundary := 2 + first.nRegs
	x, err := first.Evaluate(inputs[:boundary])
	if err != nil {
		t.Fatal(err)
	}
	yInputs := append([]bool(nil), inputs[:2]...)
	yInputs = append(yInputs, inputs[boundary:]...)
	y, err := second.Evaluate(yInputs)
	if err != nil {
		t.Fatal(err)
	}
	want := append([]bool(nil), x[:2]...)
	want = append(want, y[:2]...)
	want = append(want, x[2:]...)
	want = append(want, y[2:]...)
	checkQuantificationOutputs(t, out, inputs, want)
}

func appendReviewRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for s1 := int32(0); s1 < 8; s1++ {
		for s2 := int32(0); s2 < 8; s2++ {
			for r1 := int32(0); r1 <= 2; r1++ {
				for r2 := int32(0); r2 <= 2; r2++ {
					first, second := appendReviewGraph(s1, r1), appendReviewGraph(s2, r2)
					out := Gia_ManDupAppendNew(first, second)
					if Gia_ManPiNum(out) != 2 || Gia_ManPoNum(out) != 4 || out.nRegs != r1+r2 {
						t.Fatal("wrong shared PI, PO or independent register counts")
					}
					if name, spec := out.Identity(); name != "append" || spec != "append.aig" {
						t.Fatal("lost first graph identity")
					}
					fmt.Fprintf(&records, "%d %d %d %d %d %d %d", s1, s2, r1, r2, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
					for mask := 0; mask < 1<<uint(2+r1+r2); mask++ {
						inputs := make([]bool, 2+r1+r2)
						for i := range inputs {
							inputs[i] = mask&(1<<uint(i)) != 0
						}
						checkAppendReviewRow(t, out, first, second, inputs)
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
		}
	}
	return records.String()
}

func TestSequentialAppendTypedTransitionTables(t *testing.T) {
	appendReviewRecords(t)
}

func TestSequentialAppendAliasedTypedTransitionTables(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		for regs := int32(0); regs <= 2; regs++ {
			source := appendReviewGraph(signs, regs)
			out := Gia_ManDupAppendNew(source, source)
			if out.nRegs != 2*regs || Gia_ManPiNum(out) != 2 || Gia_ManPoNum(out) != 4 {
				t.Fatal("aliased append changed interface")
			}
			for mask := 0; mask < 1<<uint(2+2*regs); mask++ {
				inputs := make([]bool, 2+2*regs)
				for i := range inputs {
					inputs[i] = mask&(1<<uint(i)) != 0
				}
				checkAppendReviewRow(t, out, source, source, inputs)
			}
		}
	}
}

func TestSequentialAppendMatchesOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "append_sequential.c")
	want, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C sequential append: %v: %s", err, want)
	}
	gotRows := strings.Split(strings.TrimSpace(appendReviewRecords(t)), "\n")
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

func TestSequentialAppendSameGraphKeepsIndependentRegisters(t *testing.T) {
	g := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, r)
	Gia_ManAppendCo(g, Gia_ManAppendXorReal(g, a, r))
	Gia_ManSetRegNum(g, 1)
	out := Gia_ManDupAppendNew(g, g)
	for mask := 0; mask < 8; mask++ {
		a, r1, r2 := mask&1 != 0, mask&2 != 0, mask&4 != 0
		checkQuantificationOutputs(t, out, []bool{a, r1, r2}, []bool{r1, r2, a != r1, a != r2})
	}
}
