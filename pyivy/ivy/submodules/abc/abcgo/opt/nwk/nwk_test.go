package nwk

import (
	"bytes"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"strings"
	"testing"
)

func testNetwork(t *testing.T) *Nwk_Man_t {
	t.Helper()
	p := Nwk_ManAlloc()
	p.Output = &bytes.Buffer{}
	for i := 0; i < 4; i++ {
		Nwk_ManCreateCi(p, 0)
	}
	makeNode := func(a, b *Nwk_Obj_t, xor bool) *Nwk_Obj_t {
		o := Nwk_ManCreateNode(p, 2, 0)
		Nwk_ObjAddFanin(o, a)
		Nwk_ObjAddFanin(o, b)
		o.Function = hop.Hop_And(p.Hop, hop.Hop_IthVar(p.Hop, 0), hop.Hop_IthVar(p.Hop, 1))
		if xor {
			o.Function = hop.Hop_Exor(p.Hop, hop.Hop_IthVar(p.Hop, 0), hop.Hop_IthVar(p.Hop, 1))
		}
		return o
	}
	a := makeNode(p.Inputs[0], p.Inputs[1], true)
	b := makeNode(p.Inputs[2], p.Inputs[3], false)
	c := makeNode(a, b, false)
	for _, driver := range []*Nwk_Obj_t{c, a} {
		o := Nwk_ManCreateCo(p)
		Nwk_ObjAddFanin(o, driver)
	}
	return p
}
func truthOutputs(t *testing.T, p *Nwk_Man_t) [][]bool {
	t.Helper()
	g := Nwk_ManStrash(p)
	r := [][]bool{}
	for m := 0; m < 1<<g.InputCount(); m++ {
		inputs := make([]bool, g.InputCount())
		for i := range inputs {
			inputs[i] = m>>i&1 != 0
		}
		out, e := g.Evaluate(inputs)
		if e != nil {
			t.Fatal(e)
		}
		r = append(r, out)
	}
	return r
}
func TestGraphTraversalAndRoundTrip(t *testing.T) {
	p := testNetwork(t)
	if e := p.Check(); e != nil {
		t.Fatal(e)
	}
	if Nwk_ManLevel(p) != 2 || Nwk_ManLevelMax(p) != 2 {
		t.Fatal("levels")
	}
	if vec.Vec_PtrSize(Nwk_ManDfs(p)) != 9 {
		t.Fatal("DFS coverage")
	}
	if vec.Vec_PtrSize(Nwk_ManSupportNodes(p, p.Outputs, 1)) != 4 {
		t.Fatal("support")
	}
	root := p.Outputs[0].Fanins[0]
	if Nwk_ObjMffcLabel(root) != 2 {
		t.Fatal("MFFC sharing")
	}
	if e := p.Check(); e != nil {
		t.Fatal(e)
	}
	if Nwk_ObjFanout(p.Objects[4], 0) == Nwk_ObjFanout(p.Objects[4], 1) {
		t.Fatal("fanout accessor ignores index")
	}
	original := truthOutputs(t, p)
	g := Nwk_ManStrash(p)
	q := Nwk_ManDeriveFromAig(g)
	if !reflect.DeepEqual(original, truthOutputs(t, q)) {
		t.Fatal("AIG round-trip")
	}
	var blif bytes.Buffer
	if e := p.WriteBLIF(&blif, nil, nil); e != nil || !strings.Contains(blif.String(), ".end") {
		t.Fatal(e)
	}
	Nwk_ManBidecResyn(p, 0)
	if !reflect.DeepEqual(original, truthOutputs(t, p)) {
		t.Fatal("bi-decomposition changed truth")
	}
}
func TestSupportReductionAndDuplicates(t *testing.T) {
	p := Nwk_ManAlloc()
	a, b := Nwk_ManCreateCi(p, 0), Nwk_ManCreateCi(p, 0)
	node := Nwk_ManCreateNode(p, 3, 0)
	Nwk_ObjAddFanin(node, a)
	Nwk_ObjAddFanin(node, b)
	Nwk_ObjAddFanin(node, a)
	node.Function = hop.Hop_Exor(p.Hop, hop.Hop_IthVar(p.Hop, 0), hop.Hop_IthVar(p.Hop, 2))
	co := Nwk_ManCreateCo(p)
	Nwk_ObjAddFanin(co, node)
	before := truthOutputs(t, p)
	Nwk_ManRemoveDupFanins(p, 0)
	if !reflect.DeepEqual(before, truthOutputs(t, p)) {
		t.Fatal("duplicate simplification changed truth")
	}
	if len(co.Fanins[0].Fanins) != 0 {
		t.Fatal("constant support not eliminated")
	}
	if e := p.Check(); e != nil {
		t.Fatal(e)
	}
}
func TestTimingAndSpeedup(t *testing.T) {
	p := testNetwork(t)
	if Nwk_ManDelayTraceLut(p) != 2 || Nwk_ManVerifyTiming(p) != 1 {
		t.Fatal("unit timing")
	}
	library, e := if_.ParseLUTLibrary("pins", "1 1 1\n2 1 1 3\n")
	if e != nil {
		t.Fatal(e)
	}
	p.Library = library
	delay := Nwk_ManDelayTraceLut(p)
	if delay != 6 || Nwk_ManVerifyTiming(p) != 1 {
		t.Fatal("pin timing", delay)
	}
	before := truthOutputs(t, p)
	g := Nwk_ManSpeedup(p, 1, 30, 3, 0, 0)
	for m, expected := range before {
		inputs := make([]bool, len(p.Inputs))
		for i := range inputs {
			inputs[i] = m>>i&1 != 0
		}
		got, e := g.Evaluate(inputs)
		if e != nil || !reflect.DeepEqual(got, expected) {
			t.Fatal("speedup truth", e)
		}
	}
	if p.Library != library {
		t.Fatal("library ownership")
	}
}
func TestRetimingAndMatching(t *testing.T) {
	p := testNetwork(t)
	for _, forward := range []bool{true, false} {
		var v *Vec_Ptr_t
		if forward {
			v = Nwk_ManRetimeCutForward(p, 2, 0)
		} else {
			v = Nwk_ManRetimeCutBackward(p, 2, 0)
		}
		if vec.Vec_PtrSize(v) > 2 {
			t.Fatal("cut larger than registers")
		}
		if p.Check() != nil {
			t.Fatal("flow changed adjacency")
		}
	}
	g := Nwk_ManStrash(p)
	aig.Aig_ManSetRegNum(g, 2)
	if vec.Vec_PtrSize(Nwk_ManDeriveRetimingCut(g, 1, 0)) > 2 {
		t.Fatal("retiming bridge")
	}
	graph := Nwk_ManGraphAlloc(8)
	for _, pair := range [][2]int32{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 0}, {2, 6}, {6, 7}} {
		Nwk_ManGraphHashEdge(graph, pair[0], pair[1])
	}
	Nwk_ManGraphSolve(graph)
	used := map[int32]bool{}
	for _, id := range vec.Vec_IntArray(graph.Pairs) {
		if used[id] {
			t.Fatal("matching repeats vertex")
		}
		used[id] = true
	}
	if len(used) != 8 {
		t.Fatal("matching missed path coverage", vec.Vec_IntArray(graph.Pairs))
	}
	pairs := Nwk_ManLutMerge(p, &Nwk_LMPars_t{MaxLutSize: 3, MaxSuppSize: 4, MaxDistance: 2, MaxLevelDiff: 2, MaxFanout: 10, UseDiffSupp: true})
	if vec.Vec_IntSize(pairs)%2 != 0 {
		t.Fatal("odd pair count")
	}
}
