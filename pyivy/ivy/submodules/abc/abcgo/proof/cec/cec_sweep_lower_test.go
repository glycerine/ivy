package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec/core"
	"testing"
)

func TestLowerNativeGraphPreservesOutputsAndMapping(t *testing.T) {
	g := gia.Gia_ManStart(32)
	g.EnableMuxes()
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	xor := gia.Gia_ManAppendXorReal(g, a, b)
	mux := gia.Gia_ManAppendMuxReal(g, c, xor, a^1)
	buf := gia.Gia_ManAppendBuf(g, mux)
	gia.Gia_ManAppendCo(g, buf)
	gia.Gia_ManAppendCo(g, xor)
	q, mapping, aliases, err := cecLowerSweepGraph(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(mapping) != q.ObjectCount() || aliases == nil {
		t.Fatalf("mapping %v aliases %v", mapping, aliases)
	}
	for id := int32(0); int(id) < q.ObjectCount(); id++ {
		_, _, _, kind := q.Gate(id)
		if kind != 0 && kind != 1 && kind != 2 && kind != 6 {
			t.Fatalf("node %d kind %d", id, kind)
		}
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, e1 := g.Evaluate(in)
		got, e2 := q.Evaluate(in)
		if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("bits %d want %v got %v errors %v %v", bits, want, got, e1, e2)
		}
	}
}

func TestNativeSweepOriginalIDsAndWitnesses(t *testing.T) {
	for _, depth := range []int32{0, 1, 2} {
		g := gia.Gia_ManStart(32)
		g.EnableMuxes()
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendXorReal(g, a, b)
		y := gia.Gia_ManAppendBuf(g, x)
		m := gia.Gia_ManAppendMuxReal(g, c, x, a^1)
		gia.Gia_ManAppendCo(g, y)
		gia.Gia_ManAppendCo(g, m)
		var p Cec_ParFra_t
		Cec_ManFraSetDefaultParams(&p)
		p.fUseOrigIds = 1
		p.nDepthMax = depth
		q := Cec_ManSatSweeping(g, &p, 1)
		if p.Error != nil || q == nil {
			t.Fatalf("depth %d: %v", depth, p.Error)
		}
		model, err := core.NewModel(g)
		if err != nil {
			t.Fatal(err)
		}
		phase, err := model.Values(make([]bool, 3))
		if err != nil {
			t.Fatal(err)
		}
		pairs := vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))
		found := false
		for i := 0; i+1 < len(pairs); i += 2 {
			if pairs[i] == x>>1 && pairs[i+1] == y>>1 || pairs[i] == y>>1 && pairs[i+1] == x>>1 {
				found = true
			}
		}
		if !found {
			t.Fatalf("depth %d: missing buffer alias in %v", depth, pairs)
		}
		for bits := 0; bits < 8; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			want, e1 := g.Evaluate(in)
			got, e2 := q.Evaluate(in)
			if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
				t.Fatalf("depth %d input %d outputs %v vs %v errors %v %v", depth, bits, want, got, e1, e2)
			}
			values, e := model.Values(in)
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i+1 < len(pairs); i += 2 {
				u, v := pairs[i], pairs[i+1]
				if (values[u] != values[v]) != (phase[u] != phase[v]) {
					t.Fatalf("depth %d input %d invalid pair %d,%d", depth, bits, u, v)
				}
			}
		}
	}
}

func TestNativeOriginalEquivalenceConsumerPreservesFunction(t *testing.T) {
	g := gia.Gia_ManStart(16)
	g.EnableMuxes()
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	m := gia.Gia_ManAppendMuxReal(g, c, x, a)
	gia.Gia_ManAppendCo(g, m)
	result := Gia_ManComputeGiaEquivs(g, 100, 0)
	if result == nil {
		t.Fatal("nil result")
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, e1 := g.Evaluate(in)
		got, e2 := result.Evaluate(in)
		if e1 != nil || e2 != nil || got[0] != want[0] {
			t.Fatalf("bits %d want %v got %v errors %v %v", bits, want, got, e1, e2)
		}
	}
}

func TestNativeSweepPreservesSequentialInterface(t *testing.T) {
	g := gia.Gia_ManStart(64)
	g.EnableMuxes()
	in := make([]int32, 35)
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(g)
	}
	x := gia.Gia_ManAppendXorReal(g, in[33], in[2])
	m := gia.Gia_ManAppendMuxReal(g, in[34], x, in[0])
	gia.Gia_ManAppendCo(g, m)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 1)
	gia.Gia_ManSetConstrNum(g, 1)
	var p Cec_ParFra_t
	Cec_ManFraSetDefaultParams(&p)
	p.fUseOrigIds = 1
	p.nDepthMax = 2
	q := Cec_ManSatSweeping(g, &p, 1)
	if p.Error != nil || q == nil {
		t.Fatalf("sweep: %v", p.Error)
	}
	if gia.Gia_ManRegNum(q) != 1 || gia.Gia_ManConstrNum(q) != 1 || len(q.InputLiterals()) != 35 {
		t.Fatal("lost input, register or constraint interface")
	}
	consumer := Gia_ManComputeGiaEquivs(g, 100, 0)
	if consumer == nil || gia.Gia_ManRegNum(consumer) != 1 || gia.Gia_ManConstrNum(consumer) != 1 || len(consumer.InputLiterals()) != 35 {
		t.Fatal("original-ID consumer lost sequential interface")
	}
	for bits := 0; bits < 16; bits++ {
		inputs := make([]bool, 35)
		inputs[0] = bits&1 != 0
		inputs[2] = bits&2 != 0
		inputs[33] = bits&4 != 0
		inputs[34] = bits&8 != 0
		want, e1 := g.Evaluate(inputs)
		got, e2 := q.Evaluate(inputs)
		gotConsumer, e3 := consumer.Evaluate(inputs)
		if e1 != nil || e2 != nil || e3 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] || len(gotConsumer) != 2 || gotConsumer[0] != want[0] || gotConsumer[1] != want[1] {
			t.Fatalf("bits %d want %v got %v errors %v %v", bits, want, got, e1, e2)
		}
	}
}

func TestNativeSweepTransfersExistingClasses(t *testing.T) {
	g := gia.Gia_ManStart(24)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, gia.Gia_ManAppendXorReal(g, a, b), c)
	y := gia.Gia_ManAppendXorReal(g, a, gia.Gia_ManAppendXorReal(g, b, c))
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, y)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, y>>1, x>>1)
	gia.Gia_ObjSetNext(g, x>>1, y>>1)
	var p Cec_ParFra_t
	Cec_ManFraSetDefaultParams(&p)
	p.fUseOrigIds = 1
	q := Cec_ManSatSweeping(g, &p, 1)
	if q == nil || p.Error != nil {
		t.Fatalf("result %v error %v", q, p.Error)
	}
	if g.HasRepresentatives() {
		t.Fatal("source classes were not transferred")
	}
	pairs := vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))
	found := false
	for i := 0; i+1 < len(pairs); i += 2 {
		found = found || pairs[i] == x>>1 && pairs[i+1] == y>>1 || pairs[i] == y>>1 && pairs[i+1] == x>>1
	}
	if !found {
		t.Fatalf("missing original pair: %v", pairs)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, e1 := g.Evaluate(in)
		got, e2 := q.Evaluate(in)
		if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("bits %d original %v reduced %v errors %v %v", bits, want, got, e1, e2)
		}
	}
}

func TestNativeSweepRejectsFalseSeedClass(t *testing.T) {
	for _, circuit := range []int32{0, 1} {
		g := gia.Gia_ManStart(16)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendXorReal(g, a, b)
		y := gia.Gia_ManAppendAnd(g, a, b)
		gia.Gia_ManAppendCo(g, x)
		gia.Gia_ManAppendCo(g, y)
		g.StartRepresentatives()
		gia.Gia_ObjSetRepr(g, y>>1, x>>1)
		gia.Gia_ObjSetNext(g, x>>1, y>>1)
		var p Cec_ParFra_t
		Cec_ManFraSetDefaultParams(&p)
		p.fUseOrigIds = 1
		p.fRunCSat = circuit
		q := Cec_ManSatSweeping(g, &p, 1)
		if q == nil || p.Error != nil {
			t.Fatalf("circuit %d: %v", circuit, p.Error)
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			want, e1 := g.Evaluate(in)
			got, e2 := q.Evaluate(in)
			if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
				t.Fatalf("circuit %d bits %d want %v got %v errors %v %v", circuit, bits, want, got, e1, e2)
			}
		}
		pairs := vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))
		for i := 0; i+1 < len(pairs); i += 2 {
			if pairs[i] == x>>1 && pairs[i+1] == y>>1 || pairs[i] == y>>1 && pairs[i+1] == x>>1 {
				t.Fatalf("circuit %d: false proof %v", circuit, pairs)
			}
		}
	}
}
