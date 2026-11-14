package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func sequentialReviewCone(inputs, signs int32) *Gia_Man_t {
	g := Gia_ManStart(1)
	ci := make([]int32, inputs)
	for i := range ci {
		ci[i] = Gia_ManAppendCi(g)
	}
	g.EnableMuxes()
	x := Gia_ManAppendXorReal(g, ci[0]^(signs&1), ci[1])
	buffer := Gia_ManAppendBuf(g, x^((signs>>1)&1))
	m := Gia_ManAppendMuxReal(g, ci[2]^((signs>>2)&1), buffer, ci[0]^1)
	if inputs == 6 {
		m = Gia_ManAppendXorReal(g, m, ci[3])
		m = Gia_ManAppendMuxReal(g, ci[4], m, ci[5])
	}
	Gia_ManAppendCo(g, m)
	g.SetIdentity("cone", "cone.aig")
	return g
}

func sequentialCompositionReviewRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	check := func(out, src, cone *Gia_Man_t, mode int, signs, coneSigns, onlyRegs, count int32) {
		t.Helper()
		wantCI, wantCO, wantRegs := Gia_ManCiNum(src), Gia_ManCoNum(src), src.nRegs
		switch mode {
		case 0:
			wantCI++
		case 1:
			wantCO, wantRegs = 1, 0
		case 2, 3:
			wantCO += count
		}
		if Gia_ManCiNum(out) != wantCI || Gia_ManCoNum(out) != wantCO || out.nRegs != wantRegs || out.nConstrs != 0 {
			t.Fatalf("mode %d signs %d cone %d: incorrect composition interface", mode, signs, coneSigns)
		}
		sourceName, sourceSpec := src.Identity()
		if name, spec := out.Identity(); name != sourceName || spec != sourceSpec {
			t.Fatal("sequential composition changed source identity")
		}
		if Gia_ManHasDangling(out) != 0 {
			t.Fatal("composition retained dangling logic")
		}
		fmt.Fprintf(&records, "%d %d %d %d %d %d %d %d", mode, signs, coneSigns, onlyRegs, count, wantCI, wantCO, wantRegs)
		for mask := 0; mask < 1<<uint(wantCI); mask++ {
			inputs := make([]bool, wantCI)
			for i := range inputs {
				inputs[i] = mask&(1<<uint(i)) != 0
			}
			sourceInputs := inputs
			if mode == 0 {
				sourceInputs = inputs[1:]
			}
			values, err := src.Evaluate(sourceInputs)
			if err != nil {
				t.Fatal(err)
			}
			var want []bool
			switch mode {
			case 0:
				want = append(want, values...)
				if !inputs[0] {
					copy(want[Gia_ManPoNum(src):], sourceInputs[Gia_ManPiNum(src):])
				}
			case 1:
				current, err := cone.Evaluate(inputs[Gia_ManPiNum(src):])
				if err != nil {
					t.Fatal(err)
				}
				next, err := cone.Evaluate(values[Gia_ManPoNum(src):])
				if err != nil {
					t.Fatal(err)
				}
				want = []bool{current[0] && !next[0]}
			case 2, 3:
				coneInputs := inputs
				if onlyRegs != 0 {
					coneInputs = inputs[Gia_ManPiNum(src):]
				}
				additional, err := cone.Evaluate(coneInputs)
				if err != nil {
					t.Fatal(err)
				}
				want = append(want, values[:Gia_ManPoNum(src)]...)
				for i := int32(0); i < count; i++ {
					want = append(want, additional[0])
				}
				want = append(want, values[Gia_ManPoNum(src):]...)
			}
			checkQuantificationOutputs(t, out, inputs, want)
			got, err := out.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
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
	for signs := int32(0); signs < 8; signs++ {
		src := blockReviewGraph(signs, 3)
		check(Gia_ManDupSelf(src), src, nil, 0, signs, 0, 0, 0)
	}
	for signs := int32(0); signs < 8; signs++ {
		for coneSigns := int32(0); coneSigns < 8; coneSigns++ {
			src := blockReviewGraph(signs, 3)
			inv := sequentialReviewCone(3, coneSigns)
			check(Gia_ManDupInvMiter(src, inv), src, inv, 1, signs, coneSigns, 0, 0)
		}
	}
	for signs := int32(0); signs < 8; signs++ {
		for coneSigns := int32(0); coneSigns < 8; coneSigns++ {
			for onlyRegs := int32(0); onlyRegs <= 1; onlyRegs++ {
				for count := int32(0); count <= 2; count++ {
					src := blockReviewGraph(signs, 3)
					inputs := int32(6)
					if onlyRegs != 0 {
						inputs = 3
					}
					cone := sequentialReviewCone(inputs, coneSigns)
					out := Gia_ManDupAppendCones(src, []*Gia_Man_t{cone, cone, nil}, count, onlyRegs)
					check(out, src, cone, 2, signs, coneSigns, onlyRegs, count)
				}
			}
		}
	}
	for signs := int32(0); signs < 8; signs++ {
		src := sequentialReviewCone(3, signs)
		out := Gia_ManDupAppendCones(src, []*Gia_Man_t{src, src}, 2, 0)
		check(out, src, src, 3, signs, signs, 0, 2)
	}
	return records.String()
}

func TestSequentialCompositionTypedTransitions(t *testing.T) {
	sequentialCompositionReviewRecords(t)
}

func TestRegisterHoldCompositionAcrossCycles(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		src := blockReviewGraph(signs, 3)
		out := Gia_ManDupSelf(src)
		for sequence := 0; sequence < 64; sequence++ {
			state := []bool{sequence&8 != 0, sequence&16 != 0, sequence&32 != 0}
			copiedState := slices.Clone(state)
			for frame := 0; frame < 8; frame++ {
				pattern := (sequence + 3*frame) & 7
				pis := []bool{pattern&1 != 0, pattern&2 != 0, pattern&4 != 0}
				control := sequence&(1<<uint(frame%6)) != 0
				want, err := src.Evaluate(append(slices.Clone(pis), state...))
				if err != nil {
					t.Fatal(err)
				}
				if !control {
					copy(want[3:], state)
				}
				inputs := append([]bool{control}, pis...)
				inputs = append(inputs, copiedState...)
				got, err := out.Evaluate(inputs)
				if err != nil || !slices.Equal(got, want) {
					t.Fatalf("signs %d sequence %d frame %d: got %v (%v), want %v", signs, sequence, frame, got, err, want)
				}
				state, copiedState = want[3:], got[3:]
			}
		}
	}
}

func TestInductiveMiterConstantInvariants(t *testing.T) {
	for constant := int32(0); constant <= 1; constant++ {
		src := blockReviewGraph(5, 3)
		inv := Gia_ManStart(1)
		for i := 0; i < 3; i++ {
			Gia_ManAppendCi(inv)
		}
		Gia_ManAppendCo(inv, constant)
		out := Gia_ManDupInvMiter(src, inv)
		for mask := 0; mask < 64; mask++ {
			inputs := make([]bool, 6)
			for i := range inputs {
				inputs[i] = mask&(1<<uint(i)) != 0
			}
			checkQuantificationOutputs(t, out, inputs, []bool{false})
		}
	}
}

func TestSequentialCompositionMatchesOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "sequential_composition.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C sequential composition: %v: %s", err, want)
	}
	gotRows := strings.Split(strings.TrimSpace(sequentialCompositionReviewRecords(t)), "\n")
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

func TestAppendConesAcceptsSourceAsAdditionalCone(t *testing.T) {
	g := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
	out := Gia_ManDupAppendCones(g, []*Gia_Man_t{g}, 1, 0)
	for mask := 0; mask < 4; mask++ {
		a, b := mask&1 != 0, mask&2 != 0
		checkQuantificationOutputs(t, out, []bool{a, b}, []bool{a && b, a && b})
	}
}
