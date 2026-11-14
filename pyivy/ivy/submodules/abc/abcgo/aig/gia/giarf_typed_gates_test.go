package gia

import (
	"fmt"
	"testing"
)

func TestGiarfTypedGateSimulation(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		for _, recursive := range []bool{false, true} {
			t.Run(fmt.Sprintf("signs%d_recursive%t", signs, recursive), func(t *testing.T) {
				g := Gia_ManStart(1)
				a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
				g.EnableMuxes()
				x := Gia_ManAppendXorReal(g, a^(signs&1), b^((signs>>1)&1))
				control := Gia_ManAppendBuf(g, x^1)
				root := Gia_ManAppendMuxReal(g, control, b^((signs>>2)&1), c^1)
				p := Gia_ManEquivStart(g, 1000, 0)
				for round := uint32(0); round < 2; round++ {
					words := []uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0}
					for i, lit := range []int32{a, b, c} {
						words[i] ^= round * 0x5a5a5a5a
						Hcd_ObjSetSim(p, lit>>1, words[i])
					}
					if recursive {
						Gia_ManIncrementTravId(g)
						Gia_Resimulate_rec(p, root>>1)
					} else {
						Hcd_ManSimulateSimple(p)
					}
					for bit := uint(0); bit < 32; bit++ {
						av := words[0]>>bit&1 != 0
						bv := words[1]>>bit&1 != 0
						cv := words[2]>>bit&1 != 0
						xv := (av != (signs&1 != 0)) != (bv != (signs&2 != 0))
						mv := !xv && (bv != (signs&4 != 0)) || xv && !cv
						for i, lit := range []int32{x, control, root} {
							want := []bool{xv, !xv, mv}[i]
							if got := Hcd_ObjSim(p, lit>>1)>>bit&1 != 0; got != want {
								t.Fatalf("round %d node %d bit %d: %t, want %t", round, lit>>1, bit, got, want)
							}
						}
					}
				}
			})
		}
	}
}

func TestGiarfTypedReducedGraphPreservesRoot(t *testing.T) {
	g := Gia_ManStart(1)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	x := Gia_ManAppendXorReal(g, a, b^1)
	control := Gia_ManAppendBuf(g, x^1)
	root := Gia_ManAppendMuxReal(g, control, b, c^1)
	Gia_ManSetPhase(g)
	level := Gia_ManLevelNum(g)
	p := Gia_ManEquivStart(g, 1000, 0)
	Hcd_ManClassesCreate(p)
	var roots *Vec_Ptr_t
	reduced := Gia_GenerateReducedLevel(g, level, &roots)
	Gia_ManAppendCo(reduced, int32(Gia_ManObj(g, root>>1).Value))
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		want := in[1]
		if in[0] == in[1] {
			want = !in[2]
		}
		got, err := reduced.Evaluate(in)
		if err != nil || len(got) != 1 || got[0] != want {
			t.Fatalf("assignment %d: %v, %v; want %t", mask, got, err, want)
		}
	}
}

func TestGiarfTypedEquivalencesDoNotMergeXorAndAnd(t *testing.T) {
	for _, mini := range []int32{0, 1} {
		g := Gia_ManStart(1)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendXorReal(g, a, b)
		y := Gia_ManAppendAnd(g, a, b)
		Gia_ManAppendCo(g, x)
		Gia_ManAppendCo(g, y)
		Gia_ComputeEquivalences(g, 1000, mini, 0)
		q := Hcd_ManEquivToChoices(g, 1)
		for mask := 0; mask < 4; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0}
			got, err := q.Evaluate(in)
			if err != nil || len(got) != 2 || got[0] != (in[0] != in[1]) || got[1] != (in[0] && in[1]) {
				t.Fatalf("backend %d assignment %d: %v, %v", mini, mask, got, err)
			}
		}
	}
}
