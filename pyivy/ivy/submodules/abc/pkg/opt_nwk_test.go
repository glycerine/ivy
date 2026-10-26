package abc

// Tests for the Go port of src/opt/nwk.

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func makeNWKTwoInput(t *testing.T, truth uint64) *NWKNetwork {
	t.Helper()
	n, err := NewNWKNetwork("two", NWKLimits{})
	if err != nil {
		t.Fatal(err)
	}
	a, err := n.AddCI(context.Background(), "a")
	if err != nil {
		t.Fatal(err)
	}
	b, err := n.AddCI(context.Background(), "b")
	if err != nil {
		t.Fatal(err)
	}
	x, err := n.AddNode(context.Background(), "x", []int{a, b}, truth)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = n.AddCO(context.Background(), "y", x); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestNWKExhaustiveTruthAndStrash(t *testing.T) {
	ctx := context.Background()
	for truth := uint64(0); truth < 16; truth++ {
		n := makeNWKTwoInput(t, truth)
		if err := n.Validate(ctx); err != nil {
			t.Fatalf("truth %x validate: %v", truth, err)
		}
		aig, _, err := n.Strash(ctx)
		if err != nil {
			t.Fatalf("truth %x strash: %v", truth, err)
		}
		for a := 0; a < 4; a++ {
			in := []bool{a&1 != 0, a&2 != 0}
			got, err := n.Evaluate(ctx, in)
			if err != nil {
				t.Fatal(err)
			}
			want := truth&(1<<a) != 0
			if len(got) != 1 || got[0] != want {
				t.Fatalf("truth=%x a=%d got=%v want=%v", truth, a, got, want)
			}
			ag, err := aig.Evaluate(in)
			if err != nil || len(ag) != 1 || ag[0] != want {
				t.Fatalf("AIG truth=%x a=%d got=%v err=%v", truth, a, ag, err)
			}
		}
	}
}

func TestNWKTraversalSupportMFFCAndTiming(t *testing.T) {
	ctx := context.Background()
	n, _ := NewNWKNetwork("diamond", NWKLimits{})
	a, _ := n.AddCI(ctx, "a")
	b, _ := n.AddCI(ctx, "b")
	x, _ := n.AddNode(ctx, "x", []int{a, b}, 8)
	y, _ := n.AddNode(ctx, "y", []int{a, b}, 14)
	z, _ := n.AddNode(ctx, "z", []int{x, y}, 6)
	o, _ := n.AddCO(ctx, "o", z)
	dfs, err := n.DFS(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if dfs[len(dfs)-1] != o {
		t.Fatalf("DFS %v", dfs)
	}
	sup, err := n.Support(ctx, []int{z})
	if err != nil || !reflect.DeepEqual(sup, []int{a, b}) {
		t.Fatalf("support %v %v", sup, err)
	}
	mffc, err := n.MFFC(ctx, z)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mffc, []int{x, y, z}) {
		t.Fatalf("MFFC %v", mffc)
	}
	r, err := n.TraceTiming(ctx, NWKTimingOptions{PinDelays: []float64{1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if r.Delay != 4 || r.Arrival[z] != 4 || r.Slack[z] != 0 {
		t.Fatalf("timing %+v", r)
	}
	rev, err := n.ReverseDFS(ctx, []int{a})
	if err != nil {
		t.Fatal(err)
	}
	if len(rev) != 5 {
		t.Fatalf("reverse %v", rev)
	}
}

func TestNWKMinimumBaseDuplicateAndReplace(t *testing.T) {
	ctx := context.Background()
	n, _ := NewNWKNetwork("normalize", NWKLimits{})
	a, _ := n.AddCI(ctx, "a")
	b, _ := n.AddCI(ctx, "b")
	// Truth 0b1100 depends only on b when variables are [a,b].
	x, _ := n.AddNode(ctx, "x", []int{a, b}, 12)
	o, _ := n.AddCO(ctx, "o", x)
	before := make([]bool, 4)
	for i := 0; i < 4; i++ {
		v, _ := n.Evaluate(ctx, []bool{i&1 != 0, i&2 != 0})
		before[i] = v[0]
	}
	changed, err := n.MinimizeBase(ctx, x)
	if err != nil || !changed {
		t.Fatalf("min base %v %v", changed, err)
	}
	obj, _ := n.Object(x)
	if !reflect.DeepEqual(obj.Fanins, []int{b}) || obj.Truth != 2 {
		t.Fatalf("object %+v", obj)
	}
	for i := 0; i < 4; i++ {
		v, _ := n.Evaluate(ctx, []bool{i&1 != 0, i&2 != 0})
		if v[0] != before[i] {
			t.Fatal("minimum base changed truth")
		}
	}
	// Replacing with duplicate fanins and equality truth preserves a.
	if err = n.ReplaceNode(ctx, x, []int{a, a}, 8); err != nil {
		t.Fatal(err)
	}
	changed, err = n.RemoveDuplicateFanins(ctx, x)
	if err != nil || !changed {
		t.Fatalf("dedup %v %v", changed, err)
	}
	obj, _ = n.Object(x)
	if len(obj.Fanins) != 1 || obj.Fanins[0] != a || obj.Truth != 2 {
		t.Fatalf("dedup %+v", obj)
	}
	for i := 0; i < 4; i++ {
		v, _ := n.Evaluate(ctx, []bool{i&1 != 0, i&2 != 0})
		if v[0] != (i&1 != 0) {
			t.Fatalf("dedup i=%d got %v", i, v)
		}
	}
	_ = o
}

func TestNWKMergeMappingStatsAndInventory(t *testing.T) {
	ctx := context.Background()
	g, err := NewNWKMergeGraph(5, NWKLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range []NWKMergeEdge{{0, 1, 3}, {0, 2, 1}, {1, 3, 1}, {2, 4, 2}} {
		if err = g.AddEdge(e.A, e.B, e.Weight); err != nil {
			t.Fatal(err)
		}
	}
	got, err := g.Solve(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []NWKMergeEdge{{0, 2, 1}, {1, 3, 1}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("matching %v", got)
	}
	n := makeNWKTwoInput(t, 8)
	cuts, err := n.MapK(ctx, 2)
	if err != nil || len(cuts) != 1 || !reflect.DeepEqual(cuts[0].Leaves, []int{0, 1}) {
		t.Fatalf("cuts %v %v", cuts, err)
	}
	s, err := n.Stats(ctx)
	if err != nil || s.Nodes != 1 || s.Edges != 3 || s.Levels != 1 {
		t.Fatalf("stats %+v %v", s, err)
	}
	report, err := n.Report(ctx)
	if err != nil || !strings.Contains(report, "nodes=1") {
		t.Fatalf("report %q %v", report, err)
	}
	if len(NWKBuildSources) != 14 || NWKBuildSources[0] != "nwkAig.c" || NWKBuildSources[13] != "nwkUtil.c" {
		t.Fatalf("inventory %v", NWKBuildSources)
	}
	total := 0
	for _, count := range NWKDefinitionInventory {
		total += count
	}
	if len(NWKDefinitionInventory) != 15 || total != 193 {
		t.Fatalf("definition inventory len=%d total=%d", len(NWKDefinitionInventory), total)
	}
	for in, want := range map[string]string{"dir/a.blif": "dir/a", "x": "x", ".hidden": "", "a.b.c": "a.b"} {
		if got := NWKGenericFileName(in); got != want {
			t.Fatalf("generic %q = %q", in, got)
		}
	}
}

func TestNWKSourceTraversalOrderAndConstants(t *testing.T) {
	ctx := context.Background()
	n, err := NewNWKNetwork("source", NWKLimits{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := n.AddCI(ctx, "a")
	b, _ := n.AddCI(ctx, "b")
	unused, _ := n.AddCI(ctx, "unused")
	zero, err := n.AddNode(ctx, "zero", nil, 0)
	if err != nil {
		t.Fatalf("zero-fanin constant: %v", err)
	}
	one, err := n.AddNode(ctx, "one", nil, 1)
	if err != nil {
		t.Fatalf("zero-fanin constant: %v", err)
	}
	x, _ := n.AddNode(ctx, "x", []int{a, b}, 8)
	o, _ := n.AddCO(ctx, "o", x)
	_, _ = n.AddCO(ctx, "z", zero)
	_, _ = n.AddCO(ctx, "u", one)
	if err := n.Validate(ctx); err != nil {
		t.Fatal(err)
	}
	dfs, err := n.DFS(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(dfs) < 3 || !reflect.DeepEqual(dfs[:3], []int{a, b, unused}) {
		t.Fatalf("source DFS must begin with every CI: %v", dfs)
	}
	rev, err := n.ReverseDFS(ctx, []int{a})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rev, []int{o, x, a}) {
		t.Fatalf("source reverse postorder: %v", rev)
	}
	got, err := n.Evaluate(ctx, []bool{true, true, false})
	if err != nil || !reflect.DeepEqual(got, []bool{true, false, true}) {
		t.Fatalf("constant evaluation %v %v", got, err)
	}
	aig, _, err := n.Strash(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got, err = aig.Evaluate([]bool{true, true, false})
	if err != nil || !reflect.DeepEqual(got, []bool{true, false, true}) {
		t.Fatalf("constant strash %v %v", got, err)
	}
}

func TestNWKInvariantAndMergeResourceHardening(t *testing.T) {
	ctx := context.Background()
	n := makeNWKTwoInput(t, 8)
	n.cis = append(n.cis, n.cis[0])
	if _, err := n.Fanouts(ctx); !errors.Is(err, ErrInvariant) {
		t.Fatalf("corrupt CI index: %v", err)
	}
	n = makeNWKTwoInput(t, 8)
	n.retained++
	if err := n.Validate(ctx); !errors.Is(err, ErrInvariant) {
		t.Fatalf("corrupt retained count: %v", err)
	}
	if _, err := NewNWKMergeGraph(9, NWKLimits{MaxObjects: 10, MaxBytes: 64}); !errors.Is(err, ErrLimit) {
		t.Fatalf("merge constructor memory: %v", err)
	}
	g, err := NewNWKMergeGraph(3, NWKLimits{MaxObjects: 4, MaxEdges: 4, MaxBytes: 56, MaxOperations: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddEdge(0, 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := g.AddEdge(1, 2, 1); !errors.Is(err, ErrLimit) {
		t.Fatalf("merge retained bytes: %v", err)
	}
	g.Close()
	g.Close()
	if _, err := g.Solve(ctx); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed merge graph: %v", err)
	}
}

func TestNWKConstantReplacementAndDerivedLimits(t *testing.T) {
	ctx := context.Background()
	if _, err := NewNWKNetwork("too-long", NWKLimits{MaxText: 100, MaxBytes: 4}); !errors.Is(err, ErrLimit) {
		t.Fatalf("constructor retained bytes: %v", err)
	}
	n := makeNWKTwoInput(t, 8)
	if err := n.ReplaceNode(ctx, 2, nil, 1); err != nil {
		t.Fatalf("replace with source constant node: %v", err)
	}
	for a := 0; a < 4; a++ {
		got, err := n.Evaluate(ctx, []bool{a&1 != 0, a&2 != 0})
		if err != nil || !reflect.DeepEqual(got, []bool{true}) {
			t.Fatalf("constant replacement %d: %v %v", a, got, err)
		}
	}
	if _, err := n.TraceTiming(ctx, NWKTimingOptions{PinDelays: []float64{math.NaN()}}); !errors.Is(err, ErrDimension) {
		t.Fatalf("NaN pin delay: %v", err)
	}
	if _, err := n.TraceTiming(ctx, NWKTimingOptions{Required: math.Inf(1)}); !errors.Is(err, ErrDimension) {
		t.Fatalf("infinite required time: %v", err)
	}

	limited, _ := NewNWKNetwork("cuts", NWKLimits{MaxObjects: 16, MaxEdges: 16, MaxCuts: 1, MaxOperations: 1000, MaxBytes: 1 << 20})
	a, _ := limited.AddCI(ctx, "a")
	b, _ := limited.AddCI(ctx, "b")
	x, _ := limited.AddNode(ctx, "x", []int{a, b}, 8)
	_, _ = limited.AddNode(ctx, "y", []int{x, b}, 14)
	if _, err := limited.MapK(ctx, 2); !errors.Is(err, ErrLimit) {
		t.Fatalf("cut count limit: %v", err)
	}

	strashLimited, _ := NewNWKNetwork("strash", NWKLimits{MaxObjects: 4, MaxEdges: 4, MaxOperations: 1000, MaxBytes: 1 << 20})
	p, _ := strashLimited.AddCI(ctx, "p")
	q, _ := strashLimited.AddCI(ctx, "q")
	r, _ := strashLimited.AddNode(ctx, "r", []int{p, q}, 8)
	_, _ = strashLimited.AddCO(ctx, "o", r)
	if _, _, err := strashLimited.Strash(ctx); !errors.Is(err, ErrLimit) {
		t.Fatalf("derived AIG object limit: %v", err)
	}

	depth, _ := NewNWKNetwork("depth", NWKLimits{})
	dp, _ := depth.AddCI(ctx, "p")
	dq, _ := depth.AddCI(ctx, "q")
	connected, _ := depth.AddNode(ctx, "connected", []int{dp, dq}, 8)
	_, _ = depth.AddCO(ctx, "out", connected)
	dangling1, _ := depth.AddNode(ctx, "dangling1", []int{connected, dq}, 8)
	_, _ = depth.AddNode(ctx, "dangling2", []int{dangling1, dq}, 8)
	stats, err := depth.Stats(ctx)
	if err != nil || stats.Levels != 1 {
		t.Fatalf("source output-rooted level: %+v %v", stats, err)
	}
}

func TestNWKSequentialValidationAndUnsupportedStrash(t *testing.T) {
	ctx := context.Background()
	n, _ := NewNWKNetwork("seq", NWKLimits{})
	pi, _ := n.AddCI(ctx, "pi")
	lo, _ := n.AddCI(ctx, "lo")
	li, _ := n.AddCO(ctx, "li", pi)
	po, _ := n.AddCO(ctx, "po", lo)
	if err := n.SetSequentialPartitions(1, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := n.AddLatch(ctx, "latch"); err != nil {
		t.Fatal(err)
	}
	if err := n.Validate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, _, err := n.Strash(ctx); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("strash %v", err)
	}
	_, _ = li, po
}

func TestNWKLimitsCancellationAtomicityAndLifecycle(t *testing.T) {
	ctx := context.Background()
	n, _ := NewNWKNetwork("limits", NWKLimits{MaxObjects: 3, MaxEdges: 2, MaxBytes: 1024, MaxOperations: 100, MaxDepth: 8, MaxText: 100})
	a, _ := n.AddCI(ctx, "a")
	b, _ := n.AddCI(ctx, "b")
	if _, err := n.AddCI(ctx, "a"); err == nil {
		t.Fatal("duplicate name accepted")
	}
	if _, err := n.AddNode(ctx, "x", []int{a, b}, 8); err != nil {
		t.Fatal(err)
	}
	before := n.ObjectCount()
	if _, err := n.AddCO(ctx, "o", 2); !errors.Is(err, ErrLimit) {
		t.Fatalf("object limit %v", err)
	}
	if n.ObjectCount() != before {
		t.Fatal("failed add mutated")
	}
	cctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := n.Snapshot(cctx); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel %v", err)
	}
	if err := n.ReplaceNode(ctx, 2, []int{0, 99}, 8); err == nil {
		t.Fatal("bad replacement accepted")
	}
	o, _ := n.Object(2)
	if !reflect.DeepEqual(o.Fanins, []int{a, b}) {
		t.Fatal("replace mutated")
	}
	n.Close()
	n.Close()
	if _, err := n.Object(0); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed %v", err)
	}
}

func TestNWKCumulativeOperationBudget(t *testing.T) {
	ctx := context.Background()
	n, _ := NewNWKNetwork("budget", NWKLimits{MaxObjects: 8, MaxEdges: 8, MaxFanins: 2, MaxOperations: 10, MaxBytes: 1 << 20, MaxDepth: 8, MaxText: 100})
	a, _ := n.AddCI(ctx, "a")
	b, _ := n.AddCI(ctx, "b")
	x, _ := n.AddNode(ctx, "x", []int{a, b}, 8)
	n.AddCO(ctx, "o", x)
	if err := n.Validate(ctx); err != nil {
		t.Fatalf("validation should fit: %v", err)
	}
	if _, err := n.Evaluate(ctx, []bool{true, true}); !errors.Is(err, ErrLimit) {
		t.Fatalf("cumulative evaluation budget = %v", err)
	}
}

func TestNWKIndependentRandomModel5000(t *testing.T) {
	r := rand.New(rand.NewSource(1441))
	ctx := context.Background()
	for trial := 0; trial < 5000; trial++ {
		n, _ := NewNWKNetwork("m", NWKLimits{MaxObjects: 32, MaxEdges: 64, MaxOperations: 10000, MaxBytes: 1 << 20, MaxDepth: 32, MaxText: 1000, MaxFanins: 4})
		inputs := 2 + r.Intn(3)
		ids := make([]int, inputs)
		for i := range ids {
			ids[i], _ = n.AddCI(ctx, "i"+string(rune('a'+i)))
		}
		model := make([]struct {
			fs []int
			t  uint64
		}, 0)
		nodes := 1 + r.Intn(4)
		for j := 0; j < nodes; j++ {
			nf := 1 + r.Intn(min(3, len(ids)))
			fs := make([]int, nf)
			for k := range fs {
				fs[k] = ids[r.Intn(len(ids))]
			}
			truth := uint64(r.Intn(1 << (1 << nf)))
			id, err := n.AddNode(ctx, "n"+string(rune('a'+j)), fs, truth)
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
			model = append(model, struct {
				fs []int
				t  uint64
			}{append([]int(nil), fs...), truth})
		}
		root := ids[len(ids)-1]
		n.AddCO(ctx, "o", root)
		bits := r.Intn(1 << inputs)
		in := make([]bool, inputs)
		vals := make([]bool, len(ids))
		for i := range in {
			in[i] = bits&(1<<i) != 0
			vals[i] = in[i]
		}
		for j, m := range model {
			args := make([]bool, len(m.fs))
			for k, f := range m.fs {
				args[k] = vals[f]
			}
			vals[inputs+j] = nwkTruthBit(m.t, args)
		}
		got, err := n.Evaluate(ctx, in)
		if err != nil || len(got) != 1 || got[0] != vals[root] {
			t.Fatalf("trial %d got=%v want=%v err=%v", trial, got, vals[root], err)
		}
	}
}

func FuzzNWKGraphTruth(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 7, 15, 0x55, 0xaa, 0xffff} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		ctx := context.Background()
		n, _ := NewNWKNetwork("f", NWKLimits{MaxObjects: 16, MaxEdges: 32, MaxOperations: 10000, MaxBytes: 1 << 20, MaxDepth: 16, MaxText: 100})
		a, _ := n.AddCI(ctx, "a")
		b, _ := n.AddCI(ctx, "b")
		truth := seed & 15
		x, err := n.AddNode(ctx, "x", []int{a, b}, truth)
		if err != nil {
			t.Fatal(err)
		}
		n.AddCO(ctx, "o", x)
		for bits := 0; bits < 4; bits++ {
			got, err := n.Evaluate(ctx, []bool{bits&1 != 0, bits&2 != 0})
			if err != nil || got[0] != (truth&(1<<bits) != 0) {
				t.Fatalf("seed=%x bits=%d got=%v err=%v", seed, bits, got, err)
			}
		}
		aig, _, err := n.Strash(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for bits := 0; bits < 4; bits++ {
			got, err := aig.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
			if err != nil || got[0] != (truth&(1<<bits) != 0) {
				t.Fatalf("AIG seed=%x bits=%d got=%v err=%v", seed, bits, got, err)
			}
		}
	})
}
