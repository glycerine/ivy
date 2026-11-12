package aig

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func mffcFixture(shared bool) (*Graph, Edge, Edge, Edge, []Edge) {
	g := NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	u := g.AND(a, b)
	v := g.AND(u, c)
	root := g.AND(v, d)
	if shared {
		g.AddOutput(u)
	}
	g.AddOutput(root)
	return g, Aig_Regular(root), u, v, []Edge{a, b, c, d}
}
func mffcRefs(g *Graph) []int {
	result := make([]int, len(g.nodes))
	for i, n := range g.nodes {
		result[i] = n.Refs
	}
	return result
}
func mffcCheckRefs(t *testing.T, g *Graph, before []int) {
	t.Helper()
	if !slices.Equal(before, mffcRefs(g)) {
		t.Fatal("references not restored", before, mffcRefs(g))
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}
func mffcEqualSet(a, b []Edge) bool {
	a = append([]Edge(nil), a...)
	b = append([]Edge(nil), b...)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}
func TestMffcSupportAndDepthLimit(t *testing.T) {
	for _, shared := range []bool{false, true} {
		g, root, u, _, inputs := mffcFixture(shared)
		before := mffcRefs(g)
		var support []Edge
		want := int32(3)
		leaves := inputs
		if shared {
			want = 2
			leaves = []Edge{u, inputs[2], inputs[3]}
		}
		if got := Aig_NodeMffcSupp(g, root, 0, &support); got != want || !mffcEqualSet(support, leaves) {
			t.Fatal(got, support, leaves)
		}
		mffcCheckRefs(t, g, before)
		if got := Aig_NodeMffcSupp(g, root, 1, &support); got != 2 || !mffcEqualSet(support, []Edge{u, inputs[2], inputs[3]}) {
			t.Fatal(got, support)
		}
		mffcCheckRefs(t, g, before)
		support = append(support, root)
		if got := Aig_NodeMffcSupp(g, inputs[0], 0, &support); got != 0 || !slices.Equal(support, inputs[:1]) {
			t.Fatal(got, support)
		}
		if got := Aig_NodeMffcSupp(g, g.Const1(), 0, &support); got != 0 || len(support) != 0 {
			t.Fatal(got, support)
		}
	}
}
func TestMffcLabelsPowerAndCut(t *testing.T) {
	g, root, u, v, _ := mffcFixture(true)
	before := mffcRefs(g)
	probs := make([]float32, len(g.nodes))
	for i := range probs {
		probs[i] = .5
	}
	var power float32
	if got := Aig_NodeMffcLabel(g, root, &power, probs); got != 2 || power != 2 {
		t.Fatal(got, power)
	}
	if g.node(root).TravID != g.travID || g.node(v).TravID != g.travID || g.node(u).TravID == g.travID {
		t.Fatal("wrong MFFC labels")
	}
	mffcCheckRefs(t, g, before)
	if got := Aig_NodeMffcLabelCut(g, root, []Edge{v}); got != 1 {
		t.Fatal(got)
	}
	if g.node(root).TravID != g.travID || g.node(v).TravID == g.travID {
		t.Fatal("cut crossed leaf")
	}
	mffcCheckRefs(t, g, before)
	if got := Aig_NodeMffcLabel(g, root, nil, nil); got != 2 {
		t.Fatal(got)
	}
}
func TestMffcExtendCut(t *testing.T) {
	g, root, u, _, inputs := mffcFixture(true)
	before := mffcRefs(g)
	leaves := []Edge{u, inputs[2], inputs[3]}
	var result []Edge
	if !Aig_NodeMffcExtendCut(g, root, leaves, &result) || !mffcEqualSet(result, inputs) {
		t.Fatal(result)
	}
	mffcCheckRefs(t, g, before)
	if Aig_NodeMffcExtendCut(g, root, inputs, &result) {
		t.Fatal("expanded primary inputs")
	}
	mffcCheckRefs(t, g, before)
}
func TestMffcReconvergence(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	u := g.AND(a, b)
	left, right := g.AND(u, c), g.AND(u, c.Not())
	root := Aig_Regular(g.OR(left, right))
	g.AddOutput(root)
	before := mffcRefs(g)
	var support []Edge
	if got := Aig_NodeMffcSupp(g, root, 0, &support); got != 4 || !mffcEqualSet(support, []Edge{a, b, c}) {
		t.Fatal(got, support)
	}
	mffcCheckRefs(t, g, before)
}
func TestMffcDereferenceReferencePair(t *testing.T) {
	g, root, _, _, _ := mffcFixture(false)
	before := mffcRefs(g)
	if got := Aig_NodeDeref_rec(g, root, 0, nil, nil); got != 3 {
		t.Fatal(got)
	}
	if got := Aig_NodeRef_rec(g, root, 0); got != 3 {
		t.Fatal(got)
	}
	mffcCheckRefs(t, g, before)
}
func TestMffcInvalidInputRestoresReferences(t *testing.T) {
	g, root, _, _, inputs := mffcFixture(false)
	g.node(inputs[0]).Refs = 0
	before := mffcRefs(g)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("invalid reference count accepted")
			}
		}()
		Aig_NodeMffcSupp(g, root, 0, nil)
	}()
	if !slices.Equal(before, mffcRefs(g)) {
		t.Fatal("reference damage after panic")
	}
	for _, run := range []func(){
		func() { Aig_NodeMffcSupp(g, root.Not(), 0, nil) },
		func() { Aig_NodeMffcSupp(g, root, -1, nil) },
		func() { var power float32; Aig_NodeMffcLabel(g, root, &power, nil) },
		func() { var power float32; Aig_NodeMffcLabel(g, root, &power, []float32{.5}) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid MFFC input accepted")
				}
			}()
			run()
		}()
	}
}
func TestMffcTraversalRollover(t *testing.T) {
	g, root, _, _, _ := mffcFixture(false)
	g.travID = (1 << 30) - 1
	for i := range g.nodes {
		g.nodes[i].TravID = g.travID
	}
	if got := Aig_NodeMffcSupp(g, root, 0, nil); got != 3 || g.travID != 1 {
		t.Fatal(got, g.travID)
	}
}

func TestMffcUnaryBufferAndExplicitOutput(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	u := g.AND(a, b)
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{u, InvalidEdge}, Level: g.node(u).Level})
	output := Aig_ObjCreateCo(g, buffer)
	before := mffcRefs(g)
	var support []Edge
	if got := Aig_NodeMffcSupp(g, output, 0, &support); got != 1 || !mffcEqualSet(support, []Edge{a, b}) {
		t.Fatal(got, support)
	}
	mffcCheckRefs(t, g, before)
	if got := Aig_NodeMffcLabel(g, output, nil, nil); got != 1 {
		t.Fatal(got)
	}
	mffcCheckRefs(t, g, before)
}

func TestMffcReferenceRejectsNegativeCounts(t *testing.T) {
	g, root, _, _, _ := mffcFixture(false)
	fanin := g.node(root).Fanins[0]
	g.node(fanin).Refs = -1
	defer func() {
		if recover() == nil {
			t.Error("negative fanout reference accepted")
		}
		if g.node(fanin).Refs != -1 {
			t.Error("invalid count modified before rejection")
		}
	}()
	Aig_NodeRef_rec(g, root, 0)
}
func TestMffcRollbackLogTracksOnlyTouchedCone(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	for i := 0; i < 10000; i++ {
		g.NewInput()
	}
	saved := make(map[int]int)
	if got := mffcDeref(g, root, 0, nil, nil, saved); got != 1 {
		t.Fatal(got)
	}
	if len(saved) != 2 {
		t.Fatal("rollback log scaled with whole graph", len(saved))
	}
	if got := mffcRef(g, root, 0, true); got != 1 {
		t.Fatal(got)
	}
	mffcRestoreReferences(g, saved)
}

func TestMffcMatchesFanoutClosure(t *testing.T) {
	random := rand.New(rand.NewPCG(11, 29))
	for trial := 0; trial < 30; trial++ {
		g := NewGraph()
		pool := []Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 16; i++ {
			a, b := pool[random.IntN(len(pool))], pool[random.IntN(len(pool))]
			if random.IntN(2) == 0 {
				a = a.Not()
			}
			if random.IntN(2) == 0 {
				b = b.Not()
			}
			pool = append(pool, g.AND(a, b))
		}
		for i := 0; i < 3; i++ {
			g.AddOutput(pool[random.IntN(len(pool))])
		}
		parents := make(map[int]map[int]bool)
		external := make(map[int]bool)
		for id, n := range g.nodes {
			if mffcTerminal(&n) {
				continue
			}
			for _, fanin := range mffcFanins(&n) {
				child := fanin.NodeID()
				if parents[child] == nil {
					parents[child] = make(map[int]bool)
				}
				parents[child][id] = true
			}
		}
		for _, output := range g.outputs {
			external[output.NodeID()] = true
		}
		for root, n := range g.nodes {
			if n.Kind != And {
				continue
			}
			cone := map[int]bool{root: true}
			for changed := true; changed; {
				changed = false
				for id := range cone {
					for _, fanin := range mffcFanins(&g.nodes[id]) {
						child := fanin.NodeID()
						if cone[child] || external[child] || mffcTerminal(&g.nodes[child]) {
							continue
						}
						owned := true
						for parent := range parents[child] {
							if !cone[parent] {
								owned = false
								break
							}
						}
						if owned {
							cone[child] = true
							changed = true
						}
					}
				}
			}
			wantCount := int32(0)
			boundary := make(map[Edge]bool)
			for id := range cone {
				wantCount += mffcNodeCost(&g.nodes[id])
				for _, fanin := range mffcFanins(&g.nodes[id]) {
					if !cone[fanin.NodeID()] {
						boundary[Aig_Regular(fanin)] = true
					}
				}
			}
			before := mffcRefs(g)
			var support []Edge
			if got := Aig_NodeMffcSupp(g, Edge(root)<<1, 0, &support); got != wantCount {
				t.Fatal(trial, root, got, wantCount)
			}
			if len(support) != len(boundary) {
				t.Fatal("boundary size", trial, root, support, boundary)
			}
			for _, leaf := range support {
				if !boundary[leaf] {
					t.Fatal("unexpected support", leaf)
				}
			}
			mffcCheckRefs(t, g, before)
		}
	}
}
