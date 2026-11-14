package cec

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec/core"
	"github.com/glycerine/abcgo/proof/dch"
	"testing"
)

func TestSourceChoiceRefinementOptionsAndSoundness(t *testing.T) {
	for _, rings := range []int32{0, 1} {
		for _, csat := range []int32{0, 1} {
			g := gia.Gia_ManStart(40)
			in := make([]int32, 8)
			for i := range in {
				in[i] = gia.Gia_ManAppendCi(g)
			}
			x := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, in[0], in[1]), in[2])
			y := gia.Gia_ManAppendAnd(g, in[0], gia.Gia_ManAppendAnd(g, in[1], in[2]))
			rare := x
			for _, lit := range in[3:] {
				rare = gia.Gia_ManAppendAnd(g, rare, lit)
			}
			gia.Gia_ManAppendCo(g, x)
			gia.Gia_ManAppendCo(g, y)
			gia.Gia_ManAppendCo(g, rare)
			gia.Gia_ManAppendCo(g, x)
			var p Cec_ParChc_t
			Cec_ManChcSetDefaultParams(&p)
			p.nWords = 1
			p.nRounds = 1
			p.NBTLimit = 1000
			p.fUseRings = rings
			p.FUseCSat = csat
			q := Cec_ManChoiceComputationVec(g, 2, &p)
			if q == nil || p.Error != nil {
				t.Fatalf("rings %d csat %d: %v", rings, csat, p.Error)
			}
			if gia.Gia_ManCountChoices(q) == 0 {
				t.Fatalf("rings %d csat %d: missing associative alternative", rings, csat)
			}
			model, err := core.NewModel(g)
			if err != nil {
				t.Fatal(err)
			}
			phases, err := model.Values(make([]bool, 8))
			if err != nil {
				t.Fatal(err)
			}
			for bits := 0; bits < 256; bits++ {
				inputs := make([]bool, 8)
				for i := range inputs {
					inputs[i] = bits&(1<<i) != 0
				}
				source, e1 := g.Evaluate(inputs)
				out, e2 := q.Evaluate(inputs)
				if e1 != nil || e2 != nil || len(out) != 2 || out[0] != source[0] || out[1] != source[2] {
					t.Fatalf("rings %d csat %d input %d source %v output %v errors %v %v", rings, csat, bits, source, out, e1, e2)
				}
				values, e := model.Values(inputs)
				if e != nil {
					t.Fatal(e)
				}
				for id := int32(1); int(id) < g.ObjectCount(); id++ {
					repr := gia.Gia_ObjRepr(g, id)
					if repr != gia.GIA_VOID && (values[id] != values[repr]) != (phases[id] != phases[repr]) {
						t.Fatalf("rings %d csat %d input %d false class %d -> %d", rings, csat, bits, id, repr)
					}
				}
			}
		}
	}
}

func TestSourceChoicesRejectNativeGates(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, x)
	var p Cec_ParChc_t
	Cec_ManChcSetDefaultParams(&p)
	if q := Cec_ManChoiceComputationVec(g, 2, &p); q != nil || p.Error == nil {
		t.Fatalf("result %v error %v", q, p.Error)
	}
}

func TestSourceChoicesRejectsSnapshotsTooSmallForRegisters(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	gia.Gia_ManSetRegNum(g, 2)
	var p Cec_ParChc_t
	Cec_ManChcSetDefaultParams(&p)
	if q := Cec_ManChoiceComputationVec(g, 2, &p); q != nil || p.Error == nil {
		t.Fatalf("result %v error %v", q, p.Error)
	}
}

func TestComputeChoicesUsesSourceVectorPath(t *testing.T) {
	g, _, _ := sweepFixture()
	gia.Gia_ManAppendCo(g, g.OutputLiterals()[0])
	for _, circuit := range []bool{false, true} {
		copy := gia.Gia_ManDup(g)
		options := dch.DefaultParams()
		options.UseCSat = circuit
		q := Cec_ComputeChoices(copy, &options)
		if q == nil || q.OutputCount() != 1 || q.InputCount() != 3 {
			t.Fatalf("circuit %v invalid choice interface", circuit)
		}
		if aig.Aig_ManChoiceNum(q) == 0 {
			t.Fatalf("circuit %v missing associative choice", circuit)
		}
		for bits := 0; bits < 8; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			want, e1 := g.Evaluate(in)
			got, e2 := q.Evaluate(in)
			if e1 != nil || e2 != nil || got[0] != want[0] {
				t.Fatalf("circuit %v bits %d want %v got %v errors %v %v", circuit, bits, want, got, e1, e2)
			}
		}
	}
}

func TestSourceChoiceSnapshotsPreserveRegisterInterface(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, ro := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, ro)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, ro^1)
	gia.Gia_ManAppendCo(g, ro^1)
	gia.Gia_ManSetRegNum(g, 1)
	var p Cec_ParChc_t
	Cec_ManChcSetDefaultParams(&p)
	q := Cec_ManChoiceComputationVec(g, 2, &p)
	if q == nil || p.Error != nil || gia.Gia_ManRegNum(q) != 1 {
		t.Fatalf("choice result %v error %v", q, p.Error)
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		want, e1 := g.Evaluate(in)
		got, e2 := q.Evaluate(in)
		if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[2] {
			t.Fatalf("bits %d want %v got %v errors %v %v", bits, want, got, e1, e2)
		}
	}
}
