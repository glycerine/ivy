package abc

// Standalone Go tests for the src/base/acb port.

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func acbTestNetwork(t *testing.T, limits ACBLimits) (*ACBNetwork, int, int, int) {
	t.Helper()
	n, err := NewACBNetwork("top", limits)
	if err != nil {
		t.Fatal(err)
	}
	a, err := n.AddInput(context.Background(), "a", ACBRange{Left: 0, Right: 0})
	if err != nil {
		t.Fatal(err)
	}
	b, err := n.AddInput(context.Background(), "b", ACBRange{Left: 7, Right: 0, Signed: true})
	if err != nil {
		t.Fatal(err)
	}
	x, err := n.AddNode(context.Background(), "x", []int{a, b}, 0b0110)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := n.AddOutput(context.Background(), "y", x); err != nil {
		t.Fatal(err)
	}
	return n, a, b, x
}

func TestACBBasicsAndEvaluation(t *testing.T) {
	n, a, b, x := acbTestNetwork(t, ACBLimits{})
	defer n.Close()
	for assignment := 0; assignment < 4; assignment++ {
		got, err := n.Evaluate(context.Background(), []bool{assignment&1 != 0, assignment&2 != 0})
		if err != nil {
			t.Fatal(err)
		}
		want := (assignment&1 != 0) != (assignment&2 != 0)
		if len(got) != 1 || got[0] != want {
			t.Fatalf("assignment %d: got %v want %v", assignment, got, want)
		}
	}
	if err := n.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	support, err := n.Support(context.Background(), n.Outputs())
	if err != nil || len(support) != 2 || support[0] != a || support[1] != b {
		t.Fatalf("support=%v err=%v", support, err)
	}
	tfi, err := n.TFI(context.Background(), []int{x}, true)
	if err != nil || len(tfi) != 3 || tfi[0] != a || tfi[1] != b || tfi[2] != x {
		t.Fatalf("tfi=%v err=%v", tfi, err)
	}
	tfo, err := n.TFO(context.Background(), []int{a}, true)
	if err != nil || len(tfo) != 3 || tfo[0] != 4 || tfo[1] != x || tfo[2] != a {
		t.Fatalf("tfo=%v err=%v", tfo, err)
	}
	p, err := n.Profile(context.Background())
	if err != nil || p.Forward[x] != 1 || p.Forward[4] != 1 || p.Reverse[a] != 1 || p.Paths[x] != 2 {
		t.Fatalf("profile=%+v err=%v", p, err)
	}
	s, err := n.Stats(context.Background())
	if err != nil || s.Objects != 4 || s.Inputs != 2 || s.Outputs != 1 || s.Nodes != 1 || s.Edges != 3 || s.Depth != 1 {
		t.Fatalf("stats=%+v err=%v", s, err)
	}
	report, err := n.Report(context.Background())
	if err != nil || !strings.Contains(report, "objects=4") || !strings.Contains(report, "depth=1") {
		t.Fatalf("report=%q err=%v", report, err)
	}
}

func TestACBTruthExhaustive(t *testing.T) {
	for truth := 0; truth < 16; truth++ {
		n, err := NewACBNetwork("truth", ACBLimits{})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := n.AddInput(context.Background(), "a", ACBRange{})
		b, _ := n.AddInput(context.Background(), "b", ACBRange{})
		x, err := n.AddNode(context.Background(), "x", []int{a, b}, uint64(truth))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = n.AddOutput(context.Background(), "o", x)
		for assignment := 0; assignment < 4; assignment++ {
			got, err := n.Evaluate(context.Background(), []bool{assignment&1 != 0, assignment&2 != 0})
			if err != nil || got[0] != (truth&(1<<assignment) != 0) {
				t.Fatalf("truth=%x assignment=%d got=%v err=%v", truth, assignment, got, err)
			}
		}
		n.Close()
	}
}

func TestACBMinimizeAndDuplicateFanins(t *testing.T) {
	n, err := NewACBNetwork("reduce", ACBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := n.AddInput(context.Background(), "a", ACBRange{})
	b, _ := n.AddInput(context.Background(), "b", ACBRange{})
	x, _ := n.AddNode(context.Background(), "x", []int{a, b}, 0b1010)
	changed, err := n.MinimizeNodeBase(context.Background(), x)
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	o, _ := n.Object(x)
	if len(o.Fanins) != 1 || o.Fanins[0] != a || o.Truth != 0b10 {
		t.Fatalf("reduced=%+v", o)
	}
	d, _ := n.AddNode(context.Background(), "d", []int{a, a}, 0b1000)
	changed, err = n.RemoveDuplicateFanins(context.Background(), d)
	if err != nil || !changed {
		t.Fatalf("duplicate changed=%v err=%v", changed, err)
	}
	o, _ = n.Object(d)
	if len(o.Fanins) != 1 || o.Fanins[0] != a || o.Truth != 0b10 {
		t.Fatalf("dedup=%+v", o)
	}
}

func TestACBMFFCAndReplacementAtomicity(t *testing.T) {
	n, err := NewACBNetwork("mffc", ACBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := n.AddInput(context.Background(), "a", ACBRange{})
	b, _ := n.AddInput(context.Background(), "b", ACBRange{})
	n1, _ := n.AddNode(context.Background(), "n1", []int{a, b}, 0b1000)
	n2, _ := n.AddNode(context.Background(), "n2", []int{n1, a}, 0b0110)
	_, _ = n.AddOutput(context.Background(), "o", n2)
	mffc, err := n.MFFC(context.Background(), n2)
	if err != nil || len(mffc) != 2 || mffc[0] != n1 || mffc[1] != n2 {
		t.Fatalf("mffc=%v err=%v", mffc, err)
	}
	before, _ := n.Object(n2)
	if err := n.ReplaceNode(context.Background(), n2, []int{99}, 2); !errors.Is(err, ErrInvariant) {
		t.Fatalf("replace error=%v", err)
	}
	after, _ := n.Object(n2)
	if len(after.Fanins) != len(before.Fanins) || after.Fanins[0] != before.Fanins[0] || after.Truth != before.Truth {
		t.Fatalf("failed replacement mutated: before=%+v after=%+v", before, after)
	}
}

func TestACBOwnershipManagerSetsAndDefaults(t *testing.T) {
	n, _, _, x := acbTestNetwork(t, ACBLimits{})
	snapshot, err := n.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	snapshot[x-1].Fanins[0] = 99
	o, _ := n.Object(x)
	if o.Fanins[0] == 99 {
		t.Fatal("snapshot aliases network")
	}
	m, err := NewACBManager("design", ACBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	id, err := m.AddNetwork(context.Background(), n)
	if err != nil || id != 1 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if found, ok := m.Find("top"); !ok || found != id {
		t.Fatalf("find=%d,%v", found, ok)
	}
	if root, err := m.Root(); err != nil || root != n {
		t.Fatalf("root=%p err=%v", root, err)
	}
	m.Close()
	if _, err := n.Snapshot(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("manager did not close owned network: %v", err)
	}
	s := NewACBSetStore(ACBLimits{})
	s1, err := s.Intern(context.Background(), []int{4, 2, 4, 1})
	if err != nil {
		t.Fatal(err)
	}
	s2, err := s.Intern(context.Background(), []int{1, 2, 4})
	if err != nil || s1 != s2 {
		t.Fatalf("set ids=%d,%d err=%v", s1, s2, err)
	}
	values, _ := s.Values(s1)
	values[0] = 99
	again, _ := s.Values(s1)
	if len(again) != 3 || again[0] != 1 || again[1] != 2 || again[2] != 4 {
		t.Fatalf("set ownership=%v", again)
	}
	s3, err := s.Intern(context.Background(), []int{2, 3})
	if err != nil {
		t.Fatal(err)
	}
	union, err := s.Union(context.Background(), s1, s3)
	if err != nil {
		t.Fatal(err)
	}
	intersection, err := s.Intersection(context.Background(), s1, s3)
	if err != nil {
		t.Fatal(err)
	}
	difference, err := s.Difference(context.Background(), s1, s3)
	if err != nil {
		t.Fatal(err)
	}
	uv, _ := s.Values(union)
	iv, _ := s.Values(intersection)
	dv, _ := s.Values(difference)
	if len(uv) != 4 || uv[0] != 1 || uv[3] != 4 || len(iv) != 1 || iv[0] != 2 || len(dv) != 2 || dv[0] != 1 || dv[1] != 4 {
		t.Fatalf("union=%v intersection=%v difference=%v", uv, iv, dv)
	}
	if ok, err := s.IsSubset(intersection, union); err != nil || !ok {
		t.Fatalf("subset=%v err=%v", ok, err)
	}
	p := DefaultACBParameters()
	if p.LUTSize != 4 || p.TFOMaxLevels != 2 || p.TFIMaxLevels != 3 || p.FanoutMax != 20 || p.WindowNodeMax != 100 || !p.Area {
		t.Fatalf("defaults=%+v", p)
	}
	wantBuild := []string{"acbAbc.c", "acbAig.c", "acbCom.c", "acbFunc.c", "acbMfs.c", "acbPush.c", "acbSets.c", "acbTest.c", "acbUtil.c"}
	gotBuild := ACBModuleBuildInventory()
	if len(gotBuild) != len(wantBuild) {
		t.Fatalf("build=%v", gotBuild)
	}
	for i := range wantBuild {
		if gotBuild[i] != wantBuild[i] {
			t.Fatalf("build[%d]=%q", i, gotBuild[i])
		}
	}
	if strings.Count(ACBSourceCoverage(), "\n") != 8 {
		t.Fatal("coverage inventory incomplete")
	}
}

func TestACBMalformedLimitsCancellationAndLifecycle(t *testing.T) {
	if _, err := NewACBNetwork("", ACBLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("empty name err=%v", err)
	}
	n, err := NewACBNetwork("bounded", ACBLimits{MaxObjects: 3})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := n.AddInput(context.Background(), "a", ACBRange{})
	_, _ = n.AddInput(context.Background(), "b", ACBRange{})
	if _, err := n.AddNode(context.Background(), "x", []int{a}, 2); !errors.Is(err, ErrLimit) {
		t.Fatalf("object limit err=%v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	n2, _, _, _ := acbTestNetwork(t, ACBLimits{})
	if _, err := n2.Evaluate(cancelled, []bool{false, false}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel err=%v", err)
	}
	before, _ := n2.Snapshot()
	if _, err := n2.MinimizeNodeBase(cancelled, 3); !errors.Is(err, ErrCancelled) {
		t.Fatalf("minimize cancel err=%v", err)
	}
	after, _ := n2.Snapshot()
	if after[2].Truth != before[2].Truth || len(after[2].Fanins) != len(before[2].Fanins) {
		t.Fatal("cancelled reduction mutated network")
	}
	n2.objects[3].Fanins[0] = 4
	if err := n2.Validate(context.Background()); !errors.Is(err, ErrInvariant) {
		t.Fatalf("forward edge err=%v", err)
	}
	n2.Close()
	if _, err := n2.Snapshot(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed err=%v", err)
	}
	s := NewACBSetStore(ACBLimits{MaxObjects: 2})
	if _, err := s.Intern(context.Background(), []int{1}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Intern(context.Background(), []int{2}); !errors.Is(err, ErrLimit) {
		t.Fatalf("set limit err=%v", err)
	}
}

func TestACBSourceLevelsCriticalPathsAndRangeOverflow(t *testing.T) {
	n, err := NewACBNetwork("timing", ACBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := n.AddInput(context.Background(), "a", ACBRange{})
	b, _ := n.AddInput(context.Background(), "b", ACBRange{})
	n1, _ := n.AddNode(context.Background(), "n1", []int{a}, 0b10)
	n2, _ := n.AddNode(context.Background(), "n2", []int{n1}, 0b10)
	root, _ := n.AddNode(context.Background(), "root", []int{n2, b}, 0b1000)
	out, _ := n.AddOutput(context.Background(), "out", root)
	p, err := n.Profile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p.Maximum != 3 || p.Forward[out] != 3 || p.Reverse[a] != 3 || p.Reverse[b] != 1 {
		t.Fatalf("source levels=%+v", p)
	}
	if p.Paths[a] != 1 || p.Paths[b] != 0 || p.Paths[root] != 1 || p.Paths[out] != 1 {
		t.Fatalf("critical paths=%v", p.Paths)
	}
	maxInt := int(^uint(0) >> 1)
	if _, err := n.AddInput(context.Background(), "wide", ACBRange{Left: maxInt, Right: -maxInt - 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("overflowing range err=%v", err)
	}
	deep, _ := NewACBNetwork("deep", ACBLimits{MaxDepth: 1})
	di, _ := deep.AddInput(context.Background(), "di", ACBRange{})
	d1, _ := deep.AddNode(context.Background(), "d1", []int{di}, 0b10)
	d2, _ := deep.AddNode(context.Background(), "d2", []int{d1}, 0b10)
	if _, err := deep.TFI(context.Background(), []int{d2}, true); !errors.Is(err, ErrLimit) {
		t.Fatalf("depth limit err=%v", err)
	}
}

func TestACBManagerAndSetAggregateLimits(t *testing.T) {
	makeNetwork := func(name string) *ACBNetwork {
		n, err := NewACBNetwork(name, ACBLimits{})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := n.AddInput(context.Background(), name+"_a", ACBRange{})
		x, _ := n.AddNode(context.Background(), name+"_x", []int{a}, 0b10)
		_, _ = n.AddOutput(context.Background(), name+"_o", x)
		return n
	}
	m, err := NewACBManager("bounded", ACBLimits{MaxObjects: 5})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddNetwork(context.Background(), makeNetwork("one")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddNetwork(context.Background(), makeNetwork("two")); !errors.Is(err, ErrLimit) {
		t.Fatalf("aggregate manager limit err=%v", err)
	}
	if _, ok := m.Find("two"); ok {
		t.Fatal("failed manager addition was not atomic")
	}

	s := NewACBSetStore(ACBLimits{MaxBytes: 80})
	first, err := s.Intern(context.Background(), []int{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Intern(context.Background(), []int{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Intern(context.Background(), []int{7, 8, 9, 10, 11}); !errors.Is(err, ErrLimit) {
		t.Fatalf("aggregate set allocation err=%v", err)
	}
	values, _ := s.Values(first)
	if len(values) != 3 {
		t.Fatalf("failed set intern mutated store: %v", values)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.IsSubsetContext(cancelled, first, first); !errors.Is(err, ErrCancelled) {
		t.Fatalf("subset cancellation err=%v", err)
	}
	limited := NewACBSetStore(ACBLimits{MaxOperations: 16})
	if _, err := limited.Intern(context.Background(), []int{8, 7, 6, 5, 4, 3, 2, 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("sort work limit err=%v", err)
	}
}

type acbTestRNG struct{ x uint64 }

func (r *acbTestRNG) next() uint64 {
	r.x ^= r.x << 13
	r.x ^= r.x >> 7
	r.x ^= r.x << 17
	return r.x
}

func TestACBIndependentModel5000(t *testing.T) {
	rng := acbTestRNG{x: 0x9e3779b97f4a7c15}
	for trial := 0; trial < 5000; trial++ {
		inputs := 1 + int(rng.next()%4)
		nodes := 1 + int(rng.next()%5)
		n, err := NewACBNetwork("model", ACBLimits{})
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < inputs; i++ {
			if _, err := n.AddInput(context.Background(), fmtTestName("i", i), ACBRange{}); err != nil {
				t.Fatal(err)
			}
		}
		for i := 0; i < nodes; i++ {
			available := len(n.objects) - 1
			faninCount := 1 + int(rng.next()%uint64(minACB(available, 3)))
			fanins := make([]int, faninCount)
			for k := range fanins {
				fanins[k] = 1 + int(rng.next()%uint64(available))
			}
			truth := rng.next() & acbTruthMask(faninCount)
			if _, err := n.AddNode(context.Background(), fmtTestName("n", i), fanins, truth); err != nil {
				t.Fatal(err)
			}
		}
		driver := len(n.objects) - 1
		if _, err := n.AddOutput(context.Background(), "o", driver); err != nil {
			t.Fatal(err)
		}
		assignment := int(rng.next() & uint64((1<<inputs)-1))
		in := make([]bool, inputs)
		model := make([]bool, len(n.objects))
		for i := range in {
			in[i] = assignment&(1<<i) != 0
			model[i+1] = in[i]
		}
		for id := inputs + 1; id < len(n.objects); id++ {
			o := n.objects[id]
			if o.Type == ACBObjectOutput {
				model[id] = model[o.Fanins[0]]
				continue
			}
			index := 0
			for k, f := range o.Fanins {
				if model[f] {
					index |= 1 << k
				}
			}
			model[id] = o.Truth&(uint64(1)<<index) != 0
		}
		got, err := n.Evaluate(context.Background(), in)
		if err != nil || len(got) != 1 || got[0] != model[len(model)-1] {
			t.Fatalf("trial=%d got=%v model=%v err=%v", trial, got, model[len(model)-1], err)
		}
		n.Close()
	}
}

func fmtTestName(prefix string, value int) string {
	const digits = "0123456789"
	if value < 10 {
		return prefix + string(digits[value])
	}
	return prefix + "x"
}

func minACB(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func FuzzACBNetworkModel(f *testing.F) {
	f.Add(uint64(1), uint8(2), uint8(2))
	f.Add(uint64(0x96), uint8(3), uint8(4))
	f.Fuzz(func(t *testing.T, seed uint64, inputCountByte, nodeCountByte uint8) {
		inputCount := 1 + int(inputCountByte%4)
		nodeCount := 1 + int(nodeCountByte%6)
		rng := acbTestRNG{x: seed | 1}
		n, err := NewACBNetwork("fuzz", ACBLimits{MaxObjects: 32, MaxEdges: 128, MaxOperations: 1 << 16, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < inputCount; i++ {
			if _, err := n.AddInput(context.Background(), fmtTestName("i", i), ACBRange{}); err != nil {
				t.Fatal(err)
			}
		}
		for i := 0; i < nodeCount; i++ {
			available := len(n.objects) - 1
			count := 1 + int(rng.next()%uint64(minACB(available, 4)))
			fanins := make([]int, count)
			for k := range fanins {
				fanins[k] = 1 + int(rng.next()%uint64(available))
			}
			if _, err := n.AddNode(context.Background(), fmtTestName("n", i), fanins, rng.next()); err != nil {
				t.Fatal(err)
			}
		}
		_, _ = n.AddOutput(context.Background(), "o", len(n.objects)-1)
		assignment := int(rng.next() & uint64((1<<inputCount)-1))
		inputs := make([]bool, inputCount)
		for i := range inputs {
			inputs[i] = assignment&(1<<i) != 0
		}
		if _, err := n.Evaluate(context.Background(), inputs); err != nil {
			t.Fatal(err)
		}
		if _, err := n.Profile(context.Background()); err != nil {
			t.Fatal(err)
		}
		if _, err := n.Support(context.Background(), n.Outputs()); err != nil {
			t.Fatal(err)
		}
	})
}
