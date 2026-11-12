package fpga

import (
	"bytes"
	"context"
	"github.com/glycerine/abcgo/aig/hop"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func manager(t *testing.T, n, k int) *Fpga_Man_t {
	t.Helper()
	p, e := NewManager(n, 1, DefaultLibrary(k))
	if e != nil {
		t.Fatal(e)
	}
	p.Output = &bytes.Buffer{}
	return p
}
func assertEquivalent(t *testing.T, p *Fpga_Man_t) {
	t.Helper()
	for m := 0; m < 1<<len(p.Inputs); m++ {
		inputs := make([]bool, len(p.Inputs))
		for i := range inputs {
			inputs[i] = m>>i&1 != 0
		}
		a, e := p.Evaluate(inputs, false)
		if e != nil {
			t.Fatal(e)
		}
		b, e := p.Evaluate(inputs, true)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("assignment %d: original%v mapped%v", m, a, b)
		}
	}
	if area := Fpga_MappingArea(p); abs(area-p.AreaGlobal) > .001 {
		t.Fatalf("area counters: %g vs%g", area, p.AreaGlobal)
	}
}
func TestMappingRandomTruthAndRepeat(t *testing.T) {
	for k := 2; k <= 6; k++ {
		for seed := int64(0); seed < 12; seed++ {
			p := manager(t, 6, k)
			r := rand.New(rand.NewSource(seed))
			pool := append([]*Fpga_Node_t{p.Const1}, p.Inputs...)
			for i := 0; i < 60; i++ {
				a, b := pool[r.Intn(len(pool))], pool[r.Intn(len(pool))]
				a = Fpga_NotCond(a, int32(r.Intn(2)))
				b = Fpga_NotCond(b, int32(r.Intn(2)))
				pool = append(pool, Fpga_NodeAnd(p, a, b))
			}
			p.Outputs = []*Fpga_Node_t{pool[len(pool)-1], Fpga_Not(pool[len(pool)-2]), p.Const1, Fpga_Not(p.Const1), p.Inputs[0]}
			for _, switching := range []bool{false, true} {
				p.Switching = switching
				if e := p.Map(); e != nil {
					t.Fatal(e)
				}
				assertEquivalent(t, p)
			}
		}
	}
}
func TestKnownLUTCovers(t *testing.T) {
	for _, tc := range []struct {
		n, k, luts int
		delay      float32
	}{{6, 2, 5, 5}, {6, 3, 3, 3}, {6, 4, 2, 2}, {6, 6, 1, 1}} {
		p := manager(t, tc.n, tc.k)
		root := p.Inputs[0]
		for _, n := range p.Inputs[1:] {
			root = Fpga_NodeAnd(p, root, n)
		}
		p.Outputs[0] = root
		if e := p.Map(); e != nil {
			t.Fatal(e)
		}
		assertEquivalent(t, p)
		if p.Stats.LUTs != tc.luts || p.Stats.Delay != tc.delay {
			t.Fatalf("k%d: %+v", tc.k, p.Stats)
		}
	}
}
func TestChoicesAndPhase(t *testing.T) {
	p := manager(t, 4, 3)
	a, b, c, d := p.Inputs[0], p.Inputs[1], p.Inputs[2], p.Inputs[3]
	root := Fpga_NodeAnd(p, Fpga_NodeAnd(p, Fpga_NodeAnd(p, a, b), c), d)
	alt := Fpga_NodeAnd(p, Fpga_NodeAnd(p, a, b), Fpga_NodeAnd(p, c, d))
	Fpga_NodeSetChoice(p, root, alt)
	p.Outputs[0] = root
	if e := p.Map(); e != nil {
		t.Fatal(e)
	}
	assertEquivalent(t, p)
	if p.Stats.Delay != 2 || p.Choices != 1 {
		t.Fatalf("choice stats %+v", p.Stats)
	} // Also consume the secondary signal through an ordinary fanout.
	p.Outputs = append(p.Outputs, Fpga_NodeOr(p, alt, a))
	if e := p.Map(); e != nil {
		t.Fatal(e)
	}
	assertEquivalent(t, p)
	q := manager(t, 3, 3)
	a, b, c = q.Inputs[0], q.Inputs[1], q.Inputs[2]
	positive := Fpga_NodeAnd(q, a, Fpga_NodeOr(q, b, c))
	negative := Fpga_Regular(Fpga_NodeOr(q, Fpga_NodeAnd(q, a, b), Fpga_NodeAnd(q, a, c)))
	Fpga_NodeSetChoice(q, positive, negative)
	q.Outputs[0] = positive
	if e := q.Map(); e != nil {
		t.Fatal(e)
	}
	assertEquivalent(t, q)
}
func TestVariablePinTimingAndRequired(t *testing.T) {
	p := manager(t, 3, 3)
	p.Library.Delays[3] = []float32{1, 3, 5}
	p.InputArrivals = []float32{10, 3, 0}
	p.Outputs[0] = Fpga_NodeAnd(p, p.Inputs[0], Fpga_NodeAnd(p, p.Inputs[1], p.Inputs[2]))
	if e := p.Map(); e != nil {
		t.Fatal(e)
	}
	if p.Stats.Delay != 11 {
		t.Fatalf("delay %g", p.Stats.Delay)
	}
	assertEquivalent(t, p)
	p.DelayTarget = 15
	if e := p.Map(); e != nil {
		t.Fatal(e)
	}
	if p.RequiredGlobal != 15 {
		t.Fatal(p.RequiredGlobal)
	}
	for _, leaf := range p.Inputs {
		if leaf.Required < leaf.Best.Arrival {
			t.Fatalf("required %g <arrival%g", leaf.Required, leaf.Best.Arrival)
		}
	}
}
func TestLibraryParsingAndOwnership(t *testing.T) {
	p, e := ReadLibrary(strings.NewReader("# lut\n1 1 2\n2 2.5 1 3 # pins\n"), "test")
	if e != nil {
		t.Fatal(e)
	}
	q := Fpga_LutLibDup(p)
	q.Delays[2][0] = 99
	if p.Delays[2][0] != 1 {
		t.Fatal("alias")
	}
	for _, s := range []string{"", "2 1 1", "1 -1 1", "1 1 NaN", "1 1 1 2"} {
		if _, e := ReadLibrary(strings.NewReader(s), "bad"); e == nil {
			t.Fatalf("accepted %q", s)
		}
	}
	if _, e := NewManager(1, 1, DefaultLibrary(7)); e == nil {
		t.Fatal("unsupported cuts")
	}
	var out bytes.Buffer
	if e := p.Write(&out); e != nil {
		t.Fatal(e)
	}
	r, e := ReadLibrary(&out, "copy")
	if e != nil || !reflect.DeepEqual(r.Areas, p.Areas) || !reflect.DeepEqual(r.Delays, p.Delays) {
		t.Fatal(e, r)
	}
}
func TestCutReferencesAndCancellation(t *testing.T) {
	p := manager(t, 5, 2)
	shared := Fpga_NodeAnd(p, p.Inputs[0], p.Inputs[1])
	p.Outputs = []*Fpga_Node_t{Fpga_NodeAnd(p, shared, p.Inputs[2]), Fpga_NodeAnd(p, shared, p.Inputs[3]), shared}
	if e := p.Map(); e != nil {
		t.Fatal(e)
	}
	counts := make([]int32, len(p.Nodes))
	for i, n := range p.Nodes {
		counts[i] = n.Refs
	}
	for _, n := range p.Mapping.Nodes {
		Fpga_CutGetAreaRefed(p, n.Best)
	}
	for i, n := range p.Nodes {
		if counts[i] != n.Refs {
			t.Fatal("reference mutation")
		}
	}
	assertEquivalent(t, p)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p.Context = ctx
	if e := p.Map(); e != context.Canceled {
		t.Fatalf("cancellation %v", e)
	}
}
func TestCutBooleanBuilder(t *testing.T) {
	p := manager(t, 3, 3)
	p.Outputs[0] = Fpga_NodeMux(p, p.Inputs[0], Fpga_Not(p.Inputs[1]), p.Inputs[2])
	if e := p.Map(); e != nil {
		t.Fatal(e)
	}
	h := hop.Hop_ManStart()
	result := Fpga_TruthsCutBdd(h, Fpga_Regular(p.Outputs[0]).Best).(*hop.Hop_Obj_t)
	if result == nil {
		t.Fatal("missing Boolean root")
	} // Full truth is checked through Hop's public evaluator.
	truth := hop.Hop_ManComputeTruth6(h, result, int32(len(Fpga_Regular(p.Outputs[0]).Best.Leaves)))
	if uint64(truth)&truthMask(len(Fpga_Regular(p.Outputs[0]).Best.Leaves)) != Fpga_Regular(p.Outputs[0]).Best.Truth {
		t.Fatal("builder truth mismatch")
	}

	if Fpga_CutVolume(Fpga_Regular(p.Outputs[0]).Best) == 0 {
		t.Fatal("empty cut volume")
	}
}

func TestUtilitiesAndLimitedDFS(t *testing.T) {
	p := manager(t, 4, 2)
	a := Fpga_NodeAnd(p, p.Inputs[0], p.Inputs[1])
	b := Fpga_NodeAnd(p, a, p.Inputs[2])
	p.Outputs[0] = Fpga_NodeAnd(p, b, p.Inputs[3])
	if e := p.Map(); e != nil {
		t.Fatal(e)
	}
	v := Fpga_DfsLim(p, p.Outputs[0], 1)
	if len(v.Nodes) != 1 || v.Nodes[0] != p.Outputs[0] {
		t.Fatal("depth one", v.Nodes)
	}
	v = Fpga_DfsLim(p, p.Outputs[0], 2)
	if len(v.Nodes) != 3 || v.Nodes[len(v.Nodes)-1] != p.Outputs[0] {
		t.Fatal("postorder depth two")
	}
	v = Fpga_MappingLevelize(p, p.Ands)
	for level, head := range v.Nodes {
		for n := head; n != nil; n = n.NextLevel {
			if int(n.Level) != level {
				t.Fatal("level links")
			}
		}
	}
	tfo := Fpga_CollectNodeTfo(p, a)
	if len(tfo.Nodes) != 3 || tfo.Nodes[0] != a {
		t.Fatal("TFO", tfo.Nodes)
	}
	counts := Fpga_NodeGetFanoutNum(a)
	Fpga_NodeRemoveFaninFanout(a, b)
	if Fpga_NodeGetFanoutNum(a) != counts-1 {
		t.Fatal("remove")
	}
	Fpga_NodeAddFaninFanout(a, b)
	Fpga_TableResize(p)
	if Fpga_NodeAnd(p, p.Inputs[1], p.Inputs[0]) != a {
		t.Fatal("hash resize")
	}
	if Fpga_ManCheckConsistency(p) != 1 {
		t.Fatal("consistency")
	}
	if Fpga_MappingPostProcess(p) != 1 {
		t.Fatal("postprocess")
	}
	assertEquivalent(t, p)
}
func TestEmptyAndInvalidInterfaces(t *testing.T) {
	p, e := NewManager(0, 2, nil)
	if e != nil {
		t.Fatal(e)
	}
	p.Outputs = []*Fpga_Node_t{p.Const1, Fpga_Not(p.Const1)}
	if e := p.Map(); e != nil {
		t.Fatal(e)
	}
	assertEquivalent(t, p)
	q := manager(t, 2, 1)
	q.Outputs[0] = Fpga_NodeAnd(q, q.Inputs[0], q.Inputs[1])
	if e := q.Map(); e == nil {
		t.Fatal("one-input LUT mapped AND")
	}
	q.Outputs[0] = nil
	if e := q.Map(); e == nil {
		t.Fatal("unset output")
	}
	if _, e := NewManager(-1, 0, nil); e == nil {
		t.Fatal("negative count")
	}
}

func TestNegativeArrivalAndComplementedCut(t *testing.T) {
	p := manager(t, 2, 2)
	p.InputArrivals = []float32{-5, -3}
	p.Outputs[0] = Fpga_NodeAnd(p, p.Inputs[0], p.Inputs[1])
	if e := p.Map(); e != nil {
		t.Fatal(e)
	}
	if p.Stats.Delay != -2 {
		t.Fatalf("negative reference arrival: %g", p.Stats.Delay)
	}
	c := Fpga_Regular(p.Outputs[0]).Best
	h := hop.Hop_ManStart()
	r := Fpga_TruthsCutBdd(h, Fpga_CutNot(c)).(*hop.Hop_Obj_t)
	truth := hop.Hop_ManComputeTruth6(h, r, 2)
	if uint64(truth)&15 != c.Truth^15 {
		t.Fatal("lost cut complement")
	}
	p.InputArrivals = nil
	if e := p.Map(); e == nil {
		t.Fatal("missing arrivals accepted")
	}
}
