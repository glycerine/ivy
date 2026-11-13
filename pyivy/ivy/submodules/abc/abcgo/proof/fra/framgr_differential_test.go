package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/framgr"
	"math/rand"
	"slices"
	"testing"
)

// Both entry points must accept the same configuration object without a copy
// or field-by-field conversion as the manager types are consolidated.
var _ *framgr.CombOptions = (*Fra_Par_t)(nil)
var _ *framgr.Simulation = (*fraSimulation)(nil)
var _ *framgr.ClassPartition = (*classPartition)(nil)
var _ *framgr.BMC = (*Fra_Bmc_t)(nil)
var _ *framgr.Manager = (*Fra_Man_t)(nil)

// The extracted manager must preserve the original manager's combinational
// behavior under the exact configuration used by the legacy CEC fallback.
func TestSharedCECSweepMatchesOriginalManager(t *testing.T) {
	random := rand.New(rand.NewSource(94137))
	for trial := 0; trial < 50; trial++ {
		source := aig.NewGraph()
		nodes := []aig.Edge{source.NewInput(), source.NewInput(), source.NewInput(), source.NewInput()}
		for i := 0; i < 24; i++ {
			a, b := nodes[random.Intn(len(nodes))], nodes[random.Intn(len(nodes))]
			if random.Intn(2) == 0 {
				a = a.Not()
			}
			if random.Intn(2) == 0 {
				b = b.Not()
			}
			if random.Intn(2) == 0 {
				nodes = append(nodes, source.AND(a, b))
			} else {
				nodes = append(nodes, source.XOR(a, b))
			}
		}
		for i := 0; i < 3; i++ {
			edge := nodes[4+random.Intn(24)]
			if random.Intn(2) == 0 {
				edge = edge.Not()
			}
			source.AddOutput(edge)
		}
		original, shared := aig.Aig_ManDupSimple(source), aig.Aig_ManDupSimple(source)
		var pars Fra_Par_t
		Fra_ParamsDefault(&pars)
		pars.FDontShowBar, pars.FProve = 1, 1
		pars.NBTLimitNode, pars.NBTLimitMiter = 2, 300
		old, _ := fraFraigPerformMapped(original, &pars, nil)
		got := framgr.SweepForCEC(shared, 2, 300)
		if old == nil || got == nil {
			t.Fatalf("trial %d missing reduced result", trial)
		}
		for mask := 0; mask < 16; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			want, err := old.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := got.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			reference, err := source.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			for i := range want {
				if want[i] != actual[i] || actual[i] != reference[i] {
					t.Fatalf("trial %d assignment %d: old=%v shared=%v source=%v", trial, mask, want, actual, reference)
				}
			}
		}
	}
}

func TestSharedCECSweepSparseSatAndUnsat(t *testing.T) {
	for _, sat := range []bool{false, true} {
		makeGraph := func() *aig.Graph {
			g := aig.NewGraph()
			g.NewInput()
			a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
			left := g.AND(a, b)
			right := g.AND(a, b.Not())
			if sat {
				right = g.AND(a, c)
			}
			g.AddOutput(g.AND(left, right))
			return g
		}
		original, shared := makeGraph(), makeGraph()
		var pars Fra_Par_t
		Fra_ParamsDefault(&pars)
		pars.FDontShowBar, pars.FProve = 1, 1
		pars.NBTLimitNode, pars.NBTLimitMiter = 2, 300
		old, _ := fraFraigPerformMapped(original, &pars, nil)
		got := framgr.SweepForCEC(shared, 2, 300)
		if old == nil || got == nil || old.InputCount() != 4 || got.InputCount() != 4 {
			t.Fatalf("sat=%t incorrect input retention", sat)
		}
		oldSat, newSat := false, false
		for mask := 0; mask < 16; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			a, e := old.Evaluate(inputs)
			if e != nil {
				t.Fatal(e)
			}
			b, e := got.Evaluate(inputs)
			if e != nil {
				t.Fatal(e)
			}
			if a[0] != b[0] {
				t.Fatalf("sat=%t input %v outputs %v %v", sat, inputs, a, b)
			}
			oldSat = oldSat || a[0]
			newSat = newSat || b[0]
		}
		if oldSat != sat || newSat != sat {
			t.Fatalf("sat=%t old=%t new=%t", sat, oldSat, newSat)
		}
	}
}

func TestGeneralSharedSweepMatchesOriginalOptions(t *testing.T) {
	rng := rand.New(rand.NewSource(113209))
	for trial := 0; trial < 20; trial++ {
		source := aig.NewGraph()
		nodes := []aig.Edge{source.NewInput(), source.NewInput(), source.NewInput()}
		for i := 0; i < 15; i++ {
			a, b := nodes[rng.Intn(len(nodes))], nodes[rng.Intn(len(nodes))]
			if rng.Intn(2) == 0 {
				a = a.Not()
			}
			if rng.Intn(2) == 0 {
				b = b.Not()
			}
			nodes = append(nodes, source.AND(a, b))
		}
		for i := 0; i < 2; i++ {
			source.AddOutput(nodes[3+rng.Intn(15)])
		}
		for mode := 0; mode < 4; mode++ {
			a, b := aig.Aig_ManDupSimple(source), aig.Aig_ManDupSimple(source)
			var p Fra_Par_t
			Fra_ParamsDefault(&p)
			p.NSimWords = 2
			p.NBTLimitNode = 10
			p.NLevelMax = 3
			p.FDontShowBar = 1
			if mode&1 != 0 {
				p.FChoicing = 1
			}
			if mode&2 != 0 {
				p.FDoSparse = 0
			}
			q := p
			old, _ := fraFraigPerformMapped(a, &p, nil)
			shared := Fra_FraigPerform(b, &q)
			if old == nil || shared == nil || old.InputCount() != shared.InputCount() || old.OutputCount() != shared.OutputCount() {
				t.Fatalf("trial %d mode %d interface mismatch", trial, mode)
			}
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				left, e := old.Evaluate(inputs)
				if e != nil {
					t.Fatal(e)
				}
				right, e := shared.Evaluate(inputs)
				if e != nil {
					t.Fatal(e)
				}
				for i := range left {
					if left[i] != right[i] {
						t.Fatalf("trial %d mode %d input %v: %v != %v", trial, mode, inputs, left, right)
					}
				}
			}
			if p != q {
				t.Fatalf("trial %d mode %d altered caller options differently", trial, mode)
			}
		}
	}
}

func TestSharedMappedSweepPreservesSourceMapping(t *testing.T) {
	source := aig.NewGraph()
	a, b, c := source.NewInput(), source.NewInput(), source.NewInput()
	x := source.AND(a, b.Not())
	y := source.AND(x, c)
	source.AddOutput(y)
	source.AddOutput(x.Not())
	oldSource, newSource := aig.Aig_ManDupSimple(source), aig.Aig_ManDupSimple(source)
	var pars Fra_Par_t
	Fra_ParamsDefault(&pars)
	pars.NBTLimitNode = 100
	pars.FDoSparse = 1
	pars.FDontShowBar = 1
	old, oldMap := fraFraigPerformMapped(oldSource, &pars, nil)
	shared, newMap := Fra_FraigEquivenceMapped(newSource, 100, 0, nil)
	if len(oldMap) != len(newMap) {
		t.Fatal("mapping size changed")
	}
	for i := range oldMap {
		if (oldMap[i] == aig.InvalidEdge) != (newMap[i] == aig.InvalidEdge) {
			t.Fatalf("mapping presence differs for %d", i)
		}
		if oldMap[i] != aig.InvalidEdge {
			old.AddOutput(oldMap[i])
			shared.AddOutput(newMap[i])
		}
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		a, e := old.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		b, e := shared.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		if len(a) != len(b) {
			t.Fatal("output size")
		}
		for i := range a {
			if a[i] != b[i] {
				t.Fatalf("input %v mapped outputs %v != %v", in, a, b)
			}
		}
	}
}

func TestSharedBMCImplicationCallbackMatchesLegacy(t *testing.T) {
	makeBmc := func() *Fra_Bmc_t {
		g := aig.NewGraph()
		x, y, q := g.NewInput(), g.NewInput(), g.NewInput()
		n := g.AND(x, y)
		g.AddOutput(q)
		g.AddOutput(n)
		aig.Aig_ManSetRegNum(g, 1)
		b := Fra_BmcStart(g, 0, 2)
		b.PAigFrames = Fra_BmcFrames(b, 0)
		b.VImps = vec.Vec_IntAlloc(4)
		for _, pair := range [][2]aig.Edge{{x, q}, {q, x}, {x, n}, {n, x}} {
			vec.Vec_IntPush(b.VImps, Fra_ImpCreate(int32(pair[0].NodeID()), int32(pair[1].NodeID())))
		}
		return b
	}
	oldB, newB := makeBmc(), makeBmc()
	var pars Fra_Par_t
	Fra_ParamsDefault(&pars)
	pars.NBTLimitNode = 1000000
	pars.FDoSparse = 1
	pars.FDontShowBar = 1
	old, mapping := fraFraigPerformMapped(oldB.PAigFrames, &pars, func(m *Fra_Man_t) { Fra_BmcFilterImplications(m, oldB) })
	shared, sharedMapping := framgr.SweepBMCImplications(newB.PAigFrames, newB.PAig, newB.NPref, newB.NDepth, newB.PObjToFrames, newB.VImps, 1000000)
	if old == nil || shared == nil {
		t.Fatal("nil result")
	}
	if !slices.Equal(vec.Vec_IntArray(oldB.VImps), vec.Vec_IntArray(newB.VImps)) {
		t.Fatalf("implications differ: %v != %v", vec.Vec_IntArray(oldB.VImps), vec.Vec_IntArray(newB.VImps))
	}
	if !slices.Equal(mapping, sharedMapping) {
		t.Fatalf("frame mappings differ: %v != %v", mapping, sharedMapping)
	}
}
