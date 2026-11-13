package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec/core"
	"testing"
)

func TestPublicSweepOriginalIDProofs(t *testing.T) {
	for _, iters := range []int32{0, 10} {
		g, a, b := sweepFixture()
		var p Cec_ParFra_t
		Cec_ManFraSetDefaultParams(&p)
		p.nItersMax = iters
		p.fUseOrigIds = 1
		reduced := Cec_ManSatSweeping(g, &p, 1)
		if p.Error != nil || reduced == nil {
			t.Fatalf("iterations %d: %v", iters, p.Error)
		}
		pairs := vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))
		found := false
		for i := 0; i+1 < len(pairs); i += 2 {
			if (pairs[i] == a && pairs[i+1] == b) || (pairs[i] == b && pairs[i+1] == a) {
				found = true
			}
		}
		if (iters > 0) != found {
			t.Fatalf("iterations %d: proved pairs %v", iters, pairs)
		}
		if g.HasRepresentatives() {
			t.Fatal("input retained internal sweep classes")
		}
	}
}

func TestDeepSweepPreservesFunctionAndEveryPublishedPair(t *testing.T) {
	for _, depth := range []int32{0, 2, 4} {
		g := gia.Gia_ManStart(64)
		in := make([]int32, 10)
		for i := range in {
			in[i] = gia.Gia_ManAppendCi(g)
		}
		x := in[0]
		for _, lit := range in[1:] {
			x = gia.Gia_ManAppendAnd(g, x, lit)
		}
		gia.Gia_ManAppendCo(g, x)
		gia.Gia_ManAppendCo(g, x^1)
		var p Cec_ParFra_t
		Cec_ManFraSetDefaultParams(&p)
		p.nWords, p.nRounds, p.nItersMax, p.nBTLimit, p.nDepthMax, p.fUseOrigIds = 1, 1, 30, 1000, depth, 1
		result := Cec_ManSatSweeping(g, &p, 1)
		if p.Error != nil || result == nil {
			t.Fatalf("depth %d: %v", depth, p.Error)
		}
		model, err := core.NewModel(g)
		if err != nil {
			t.Fatal(err)
		}
		phases, err := model.Values(make([]bool, 10))
		if err != nil {
			t.Fatal(err)
		}
		pairs := vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))
		for bits := 0; bits < 1024; bits++ {
			inputs := make([]bool, 10)
			for i := range inputs {
				inputs[i] = bits&(1<<i) != 0
			}
			want, e1 := g.Evaluate(inputs)
			got, e2 := result.Evaluate(inputs)
			if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
				t.Fatalf("depth %d input %d original %v reduced %v errors %v %v", depth, bits, want, got, e1, e2)
			}
			values, err := model.Values(inputs)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i+1 < len(pairs); i += 2 {
				a, b := pairs[i], pairs[i+1]
				if (values[a] != values[b]) != (phases[a] != phases[b]) {
					t.Fatalf("depth %d input %d invalid pair %d,%d", depth, bits, a, b)
				}
			}
		}
	}
}

func TestPublicSweepRejectsInvalidInputs(t *testing.T) {
	var p Cec_ParFra_t
	Cec_ManFraSetDefaultParams(&p)
	if got := Cec_ManSatSweeping(nil, &p, 1); got != nil || p.Error == nil {
		t.Fatalf("nil graph: %v %v", got, p.Error)
	}
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	Cec_ManFraSetDefaultParams(&p)
	p.fDualOut = 1
	if got := Cec_ManSatSweeping(g, &p, 1); got != nil || p.Error == nil {
		t.Fatalf("odd dual miter: %v %v", got, p.Error)
	}
	g = gia.Gia_ManStart(8)
	a = gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	gia.Gia_ManAppendCo(g, x)
	Cec_ManFraSetDefaultParams(&p)
	if got := Cec_ManSatSweeping(g, &p, 1); got != nil || p.Error == nil {
		t.Fatalf("native XOR must be rejected: %v %v", got, p.Error)
	}
}

func TestPublicSweepDeepOriginalIDProofs(t *testing.T) {
	for _, depth := range []int32{0, 2} {
		g, a, b := sweepFixture()
		var p Cec_ParFra_t
		Cec_ManFraSetDefaultParams(&p)
		p.fUseOrigIds, p.nDepthMax = 1, depth
		if got := Cec_ManSatSweeping(g, &p, 1); got == nil || p.Error != nil {
			t.Fatalf("depth %d: graph %v error %v", depth, got, p.Error)
		}
		pairs := vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))
		found := false
		for i := 0; i+1 < len(pairs); i += 2 {
			found = found || pairs[i] == a && pairs[i+1] == b || pairs[i] == b && pairs[i+1] == a
		}
		if !found {
			t.Fatalf("depth %d missing equivalent pair: %v", depth, pairs)
		}
	}
	g, a, b := sweepFixture()
	var p Cec_ParFra_t
	Cec_ManFraSetDefaultParams(&p)
	p.fUseOrigIds, p.fRunCSat = 1, 1
	if got := Cec_ManSatSweeping(g, &p, 1); got == nil || p.Error != nil {
		t.Fatalf("circuit SAT: graph %v error %v", got, p.Error)
	}
	pairs := vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))
	found := false
	for i := 0; i+1 < len(pairs); i += 2 {
		found = found || pairs[i] == a && pairs[i+1] == b || pairs[i] == b && pairs[i+1] == a
	}
	if !found {
		t.Fatalf("circuit SAT missing equivalent pair: %v", pairs)
	}
}
