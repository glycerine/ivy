package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
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
		q := cecTestSweepLowered(g, &p)
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

func TestNativeOriginalEquivalenceConsumerRejectsUnsupportedInput(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	gia.Gia_ManAppendCo(g, x)
	if got := Gia_ManComputeGiaEquivs(g, 100, 0); got != nil {
		t.Fatal("unsupported native XOR accepted")
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
	q := cecTestSweepLowered(g, &p)
	if p.Error != nil || q == nil {
		t.Fatalf("sweep: %v", p.Error)
	}
	if gia.Gia_ManRegNum(q) != 1 || gia.Gia_ManConstrNum(q) != 1 || len(q.InputLiterals()) != 35 {
		t.Fatal("lost input, register or constraint interface")
	}
	for bits := 0; bits < 16; bits++ {
		inputs := make([]bool, 35)
		inputs[0] = bits&1 != 0
		inputs[2] = bits&2 != 0
		inputs[33] = bits&4 != 0
		inputs[34] = bits&8 != 0
		want, e1 := g.Evaluate(inputs)
		got, e2 := q.Evaluate(inputs)
		if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
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
	q := cecTestSweepLowered(g, &p)
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
		q := cecTestSweepLowered(g, &p)
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

func TestNativeSweepComplementedSeedAlias(t *testing.T) {
	for _, track := range []int32{0, 1} {
		g := gia.Gia_ManStart(16)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendXorReal(g, a, b)
		y := gia.Gia_ManAppendBuf(g, x^1)
		gia.Gia_ManAppendCo(g, x)
		gia.Gia_ManAppendCo(g, y)
		g.StartRepresentatives()
		gia.Gia_ObjSetRepr(g, y>>1, x>>1)
		gia.Gia_ObjSetNext(g, x>>1, y>>1)
		var p Cec_ParFra_t
		Cec_ManFraSetDefaultParams(&p)
		p.fUseOrigIds = track
		q := cecTestSweepLowered(g, &p)
		if q == nil || p.Error != nil {
			t.Fatalf("track %d: %v", track, p.Error)
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			want, e1 := g.Evaluate(in)
			got, e2 := q.Evaluate(in)
			if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
				t.Fatalf("track %d bits %d want %v got %v errors %v %v", track, bits, want, got, e1, e2)
			}
		}
		if track != 0 {
			pairs := vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))
			found := false
			for i := 0; i+1 < len(pairs); i += 2 {
				found = found || pairs[i] == x>>1 && pairs[i+1] == y>>1 || pairs[i] == y>>1 && pairs[i+1] == x>>1
			}
			if !found {
				t.Fatalf("missing complemented alias %v", pairs)
			}
		}
	}
}

func TestNativeSweepConflictLimitedSeed(t *testing.T) {
	const pigeons, holes = 6, 5
	ag := aig.NewGraph()
	vars := make([][]aig.Edge, pigeons)
	for i := range vars {
		vars[i] = make([]aig.Edge, holes)
		for j := range vars[i] {
			vars[i][j] = ag.NewInput()
		}
	}
	formula := ag.Const1()
	for i := range vars {
		clause := ag.Const0()
		for _, v := range vars[i] {
			clause = ag.OR(clause, v)
		}
		formula = ag.AND(formula, clause)
	}
	for h := 0; h < holes; h++ {
		for i := 0; i < pigeons; i++ {
			for j := 0; j < i; j++ {
				formula = ag.AND(formula, ag.OR(vars[i][h].Not(), vars[j][h].Not()))
			}
		}
	}
	ag.AddOutput(formula)
	g := gia.Gia_ManFromAig(ag)
	formulaLit := g.OutputLiterals()[0]
	in := g.InputLiterals()
	native := gia.Gia_ManAppendXorReal(g, in[0], in[1])
	gia.Gia_ManAppendCo(g, native)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, formulaLit>>1, 0)
	lower, _, _, err := cecLowerSweepGraph(g)
	if err != nil {
		t.Fatal(err)
	}
	fra := Cec_ManFraStart(lower, &Cec_ParFra_t{nDepthMax: 1})
	query := Cec_ManFraSpecReduction(fra)
	if gia.Gia_ManCoNum(query) == 0 {
		t.Fatal("missing seeded query")
	}
	Cec_ManSatSolve(nil, query, &Cec_ParSat_t{nBTLimit: 1}, nil, nil, nil, 0)
	unknown := false
	for i := int32(0); i < gia.Gia_ManCoNum(query); i++ {
		o := gia.Gia_ManCo(query, i)
		unknown = unknown || gia.Gia_ObjMark0(o) == 0 && gia.Gia_ObjMark1(o) == 0
	}
	if !unknown {
		t.Fatal("fixture did not reach conflict-limited unknown")
	}
	var pars Cec_ParFra_t
	Cec_ManFraSetDefaultParams(&pars)
	pars.nBTLimit = 1
	pars.nItersMax = 1
	pars.fSatSweeping = 1
	pars.fUseOrigIds = 1
	result := cecTestSweepLowered(g, &pars)
	if result == nil || pars.Error != nil {
		t.Fatalf("sweep: %v", pars.Error)
	}
	pairs := vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i] == 0 && pairs[i+1] == formulaLit>>1 || pairs[i+1] == 0 && pairs[i] == formulaLit>>1 {
			t.Fatalf("unknown candidate published: %v", pairs)
		}
	}
	for sample := 0; sample < 128; sample++ {
		inputs := make([]bool, len(in))
		seed := uint32(sample + 1)
		for i := range inputs {
			seed = seed*1664525 + 1013904223
			inputs[i] = seed>>31 != 0
		}
		want, e1 := g.Evaluate(inputs)
		got, e2 := result.Evaluate(inputs)
		if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("sample %d want %v got %v errors %v %v", sample, want, got, e1, e2)
		}
	}
}

// This test-only helper exercises the standalone lowering utilities; source
// Cec_ManSatSweeping does not lower native nodes.
func cecTestSweepLowered(graph *Gia_Man_t, params *Cec_ParFra_t) *Gia_Man_t {
	lowered, mapping, aliases, err := cecLowerSweepGraph(graph)
	if err != nil {
		params.Error = err
		return nil
	}
	defer gia.Gia_ManStop(lowered)
	if words := graph.SimulationWordCount(); words > 0 {
		data := graph.SimulationInputs()
		if len(data) != int(gia.Gia_ManCiNum(graph)*words) {
			params.Error = fmt.Errorf("invalid simulation size")
			return nil
		}
		gia.Gia_ManCecSimAlloc(lowered, words)
		for _, word := range data {
			gia.Gia_ManCecSaveCIWord(lowered, word)
		}
	}
	result, err := cecOriginalSweepMapped(lowered, graph, mapping, aliases, params)
	params.Error = err
	if err == nil && graph.HasRepresentatives() {
		graph.ClearRepresentatives()
	}
	return result
}
