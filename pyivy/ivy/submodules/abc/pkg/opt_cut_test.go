package abc

// Tests for the port of src/opt/cut.

import (
	"errors"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

type kCutTestNode struct {
	left, right       int
	invLeft, invRight bool
}

func TestKCutMergeDominanceAndFormat(t *testing.T) {
	a := KCut{Leaves: []int{1, 4, 8}, Signature: kCutSignature([]int{1, 4, 8})}
	b := KCut{Leaves: []int{2, 4, 9}, Signature: kCutSignature([]int{2, 4, 9})}
	got, ok := MergeKCuts(a, b, 5)
	if !ok || !reflect.DeepEqual(got.Leaves, []int{1, 2, 4, 8, 9}) {
		t.Fatalf("merge = %v, %v", got.Leaves, ok)
	}
	if _, ok := MergeKCuts(a, b, 4); ok {
		t.Fatal("infeasible merge accepted")
	}
	if !kCutSubset([]int{1, 4}, got.Leaves) || kCutSubset([]int{1, 5}, got.Leaves) {
		t.Fatal("subset relation is wrong")
	}
	text, err := FormatKCut(got, false)
	if err != nil || text != "5 : { 1 2 4 8 9 }" {
		t.Fatalf("format = %q, %v", text, err)
	}
	seq := KCut{Leaves: []int{2 << 8, 3<<8 | 7}}
	text, err = FormatKCut(seq, true)
	if err != nil || text != "2 : { 2 3(7) }" {
		t.Fatalf("sequential format = %q, %v", text, err)
	}
}

func TestKCutManagerTruthAndIndependentEnumeration(t *testing.T) {
	nodes := []kCutTestNode{
		{}, {}, {},
		{0, 1, false, false},
		{1, 2, true, false},
		{3, 4, false, true},
	}
	m, err := NewKCutManager(KCutOptions{MaxLeaves: 4, KeepMax: 32, MaxNodes: len(nodes), ComputeTruth: true, Filter: true, Record: true})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := m.SetTrivial(i); err != nil {
			t.Fatal(err)
		}
	}
	want := make([][][]int, len(nodes))
	for i := 0; i < 3; i++ {
		want[i] = [][]int{{i}}
	}
	for id := 3; id < len(nodes); id++ {
		n := nodes[id]
		cuts, err := m.ComputeAND(id, n.left, n.right, n.invLeft, n.invRight, true, 0)
		if err != nil {
			t.Fatal(err)
		}
		want[id] = independentKCutSets(id, want[n.left], want[n.right], 4, 32)
		if got := kCutLeafSets(cuts); !reflect.DeepEqual(got, want[id]) {
			t.Fatalf("node %d leaves\n got %v\nwant %v", id, got, want[id])
		}
		if err := VerifyKCuts(cuts, 4, true); err != nil {
			t.Fatalf("node %d invalid: %v", id, err)
		}
		for _, cut := range cuts {
			for assignment := 0; assignment < 1<<uint(len(cut.Leaves)); assignment++ {
				got := kCutTruthBit(cut.Truth, assignment)
				wantValue := evalKCutGraph(nodes, id, cut.Leaves, assignment)
				if got != wantValue {
					t.Fatalf("node %d cut %v assignment %d: got %v want %v", id, cut.Leaves, assignment, got, wantValue)
				}
			}
		}
	}
	records, err := m.Records()
	if err != nil || len(records) == 0 {
		t.Fatalf("records = %v, %v", records, err)
	}
	records[0].Node = -1
	records2, _ := m.Records()
	if records2[0].Node < 0 {
		t.Fatal("records alias manager storage")
	}
}

func TestKCutOwnershipDropUnionCancellationAndLifecycle(t *testing.T) {
	cancel := false
	m, err := NewKCutManager(KCutOptions{
		MaxLeaves: 3, KeepMax: 8, MaxNodes: 8, ComputeTruth: true, Filter: true,
		Drop: true, Cancelled: func() bool { return cancel },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.SetFanoutCounts([]int{2, 1}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := m.SetTrivial(i); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.ComputeAND(3, 0, 1, false, false, true, 0); err != nil {
		t.Fatal(err)
	}
	before, _ := m.Cuts(3)
	cancel = true
	if _, err := m.ComputeAND(3, 1, 2, false, false, true, 0); !errors.Is(err, ErrKCutCancelled) {
		t.Fatalf("cancel error = %v", err)
	}
	cancel = false
	after, _ := m.Cuts(3)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("cancelled operation was not atomic")
	}
	after[0].Leaves[0] = 99
	again, _ := m.Cuts(3)
	if again[0].Leaves[0] == 99 {
		t.Fatal("Cuts returned borrowed storage")
	}
	if err := m.DropFanout(0); err != nil {
		t.Fatal(err)
	}
	if cuts, _ := m.Cuts(0); len(cuts) == 0 {
		t.Fatal("first fanout drop freed cuts early")
	}
	if err := m.DropFanout(0); err != nil {
		t.Fatal(err)
	}
	if cuts, _ := m.Cuts(0); len(cuts) != 0 {
		t.Fatal("last fanout drop did not free cuts")
	}
	m.Close()
	if _, err := m.Cuts(1); !errors.Is(err, ErrKCutClosed) {
		t.Fatalf("closed error = %v", err)
	}
	m.Close()
}

func TestKCutChoiceUnionAndGlobalFilter(t *testing.T) {
	if _, err := NewKCutManager(KCutOptions{MaxLeaves: 2, KeepMax: 4, MaxNodes: 5}); !errors.Is(err, ErrKCutBounds) {
		t.Fatalf("small K error = %v", err)
	}
	m, err := NewKCutManager(KCutOptions{MaxLeaves: 3, KeepMax: 8, MaxNodes: 5, Filter: true})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := m.SetTrivial(i); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.ComputeAND(3, 0, 1, false, false, true, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := m.ComputeAND(4, 0, 2, false, false, true, 0); err != nil {
		t.Fatal(err)
	}
	cuts, err := m.UnionChoices([]int{3, 4})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{3}, {0, 1}, {0, 2}}
	if got := kCutLeafSets(cuts); !reflect.DeepEqual(got, want) {
		t.Fatalf("choice union = %v want %v", got, want)
	}
	if moved, _ := m.Cuts(4); len(moved) != 0 {
		t.Fatal("choice source was not released")
	}

	global, _ := NewKCutManager(KCutOptions{MaxLeaves: 3, KeepMax: 8, MaxNodes: 4, Filter: true, GlobalOnly: true})
	if err := global.SetGlobalNodes([]bool{false, true}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := global.SetTrivial(i); err != nil {
			t.Fatal(err)
		}
	}
	cuts, err = global.ComputeAND(2, 0, 1, false, false, true, 0)
	if err != nil || len(cuts) != 2 {
		t.Fatalf("global cut = %v, %v", cuts, err)
	}
	global2, _ := NewKCutManager(KCutOptions{MaxLeaves: 3, KeepMax: 8, MaxNodes: 4, Filter: true, GlobalOnly: true})
	for i := 0; i < 2; i++ {
		if err := global2.SetTrivial(i); err != nil {
			t.Fatal(err)
		}
	}
	cuts, err = global2.ComputeAND(2, 0, 1, false, false, true, 0)
	if err != nil || len(cuts) != 1 || !reflect.DeepEqual(cuts[0].Leaves, []int{2}) {
		t.Fatalf("local-only global cut = %v, %v", cuts, err)
	}
}

func TestKCutChoiceUnionCancellationIsAtomic(t *testing.T) {
	checksUntilCancel := 0
	m, err := NewKCutManager(KCutOptions{
		MaxLeaves: 3,
		KeepMax:   8,
		MaxNodes:  8,
		Filter:    true,
		Cancelled: func() bool {
			if checksUntilCancel == 0 {
				return false
			}
			checksUntilCancel--
			return checksUntilCancel == 0
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := m.SetTrivial(i); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.ComputeAND(3, 0, 1, false, false, true, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := m.ComputeAND(4, 1, 2, false, false, true, 0); err != nil {
		t.Fatal(err)
	}
	beforeStats, _ := m.Stats()
	before3, _ := m.Cuts(3)
	before4, _ := m.Cuts(4)
	checksUntilCancel = 3 // entry, first candidate, then cancel after work began
	if _, err := m.UnionChoices([]int{3, 4}); !errors.Is(err, ErrKCutCancelled) {
		t.Fatalf("choice cancellation error = %v", err)
	}
	checksUntilCancel = 0
	afterStats, _ := m.Stats()
	after3, _ := m.Cuts(3)
	after4, _ := m.Cuts(4)
	if beforeStats != afterStats || !reflect.DeepEqual(before3, after3) || !reflect.DeepEqual(before4, after4) {
		t.Fatalf("cancelled union mutated manager: stats %v -> %v, root %v -> %v, choice %v -> %v", beforeStats, afterStats, before3, after3, before4, after4)
	}
}

func TestKCutCompositionAndMapping(t *testing.T) {
	// factor is x0 XOR x1, with leaf 5 substituted by tree x2 AND x3.
	factor := KCut{Leaves: []int{5, 9}, Truth: truthFromFunction(2, func(a int) bool { return ((a >> 0) & 1) != ((a >> 1) & 1) })}
	tree := KCut{Leaves: []int{2, 3}, Truth: truthFromFunction(2, func(a int) bool { return a&3 == 3 })}
	got, err := ComposeKCutTruth(factor, 5, tree, []int{2, 3, 9})
	if err != nil {
		t.Fatal(err)
	}
	for a := 0; a < 8; a++ {
		want := (a&3 == 3) != (a&4 != 0)
		if kCutTruthBit(got.Truth, a) != want {
			t.Fatalf("composition assignment %d", a)
		}
	}
	if !reflect.DeepEqual(factor.Leaves, []int{5, 9}) {
		t.Fatal("composition mutated factor")
	}

	m, _ := NewKCutManager(KCutOptions{MaxLeaves: 3, KeepMax: 8, MaxNodes: 6, Filter: true})
	for i := 0; i < 3; i++ {
		if err := m.SetTrivial(i); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.ComputeAND(3, 0, 1, false, false, true, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := m.ComputeAND(4, 3, 2, false, false, true, 0); err != nil {
		t.Fatal(err)
	}
	if delay, err := m.MapNode(3); err != nil || delay != 1 {
		t.Fatalf("delay3 = %d, %v", delay, err)
	}
	if delay, err := m.MapNode(4); err != nil || delay != 1 {
		t.Fatalf("delay4 = %d, %v", delay, err)
	}
	if area, err := m.MappingArea(4); err != nil || area != 1 {
		t.Fatalf("area = %d, %v", area, err)
	}
}

func TestKCutNPNCanonicalization(t *testing.T) {
	majority := KCut{Leaves: []int{0, 1, 2}, Truth: truthFromFunction(3, func(a int) bool {
		return (a&1)+((a>>1)&1)+((a>>2)&1) >= 2
	})}
	direct, negative, err := CanonicalizeKCutNPN(majority)
	if err != nil {
		t.Fatal(err)
	}
	permuted := KCut{Leaves: []int{3, 4, 5}, Truth: truthFromFunction(3, func(a int) bool {
		// Swap variables zero and two, then negate variable one.
		original := ((a >> 2) & 1) | (((a>>1)&1)^1)<<1 | (a&1)<<2
		return kCutTruthBit(majority.Truth, original)
	})}
	direct2, negative2, err := CanonicalizeKCutNPN(permuted)
	if err != nil {
		t.Fatal(err)
	}
	if direct.Truth != direct2.Truth || negative.Truth != negative2.Truth {
		t.Fatalf("canonical mismatch: %x/%x and %x/%x", direct.Truth, negative.Truth, direct2.Truth, negative2.Truth)
	}
	if len(direct.Permutation) != 3 || !negative.OutputNegated {
		t.Fatal("canonical metadata missing")
	}
}

func TestKCutRandomModelCampaign(t *testing.T) {
	rng := rand.New(rand.NewSource(0x435554))
	for run := 0; run < 5000; run++ {
		inputs := 2 + rng.Intn(3)
		nodeCount := inputs + 1 + rng.Intn(5)
		maxLeaves := 3 + rng.Intn(2)
		m, err := NewKCutManager(KCutOptions{MaxLeaves: maxLeaves, KeepMax: 32, MaxNodes: nodeCount, ComputeTruth: true, Filter: true})
		if err != nil {
			t.Fatal(err)
		}
		type edgeNode struct {
			a, b   int
			ca, cb bool
		}
		graph := make([]edgeNode, nodeCount)
		model := make([][][]int, nodeCount)
		for i := 0; i < inputs; i++ {
			if err := m.SetTrivial(i); err != nil {
				t.Fatal(err)
			}
			model[i] = [][]int{{i}}
		}
		for id := inputs; id < nodeCount; id++ {
			graph[id] = edgeNode{rng.Intn(id), rng.Intn(id), rng.Intn(2) != 0, rng.Intn(2) != 0}
			g := graph[id]
			cuts, err := m.ComputeAND(id, g.a, g.b, g.ca, g.cb, true, 0)
			if err != nil {
				t.Fatalf("run %d node %d: %v", run, id, err)
			}
			model[id] = independentKCutSets(id, model[g.a], model[g.b], maxLeaves, 32)
			if !reflect.DeepEqual(kCutLeafSets(cuts), model[id]) {
				t.Fatalf("run %d node %d: got %v want %v", run, id, kCutLeafSets(cuts), model[id])
			}
		}
	}
}

func FuzzKCutEnumeration(f *testing.F) {
	f.Add([]byte{3, 0, 1, 0, 1, 2, 3, 1, 2})
	f.Add([]byte{4, 3, 2, 1, 0, 255, 17, 9})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		inputs := 2 + int(data[0]%3)
		andCount := min(7, max(1, (len(data)-1)/3))
		nodeCount := inputs + andCount
		m, err := NewKCutManager(KCutOptions{MaxLeaves: 4, KeepMax: 48, MaxNodes: nodeCount, ComputeTruth: true, Filter: true})
		if err != nil {
			t.Fatal(err)
		}
		model := make([][][]int, nodeCount)
		for i := 0; i < inputs; i++ {
			if err := m.SetTrivial(i); err != nil {
				t.Fatal(err)
			}
			model[i] = [][]int{{i}}
		}
		for id := inputs; id < nodeCount; id++ {
			p := 1 + 3*(id-inputs)
			b0, b1, flags := byte(id), byte(id*7), byte(0)
			if p < len(data) {
				b0 = data[p]
			}
			if p+1 < len(data) {
				b1 = data[p+1]
			}
			if p+2 < len(data) {
				flags = data[p+2]
			}
			a, b := int(b0)%id, int(b1)%id
			cuts, err := m.ComputeAND(id, a, b, flags&1 != 0, flags&2 != 0, true, 0)
			if err != nil {
				t.Fatal(err)
			}
			model[id] = independentKCutSets(id, model[a], model[b], 4, 48)
			if !reflect.DeepEqual(kCutLeafSets(cuts), model[id]) {
				t.Fatalf("node %d got %v want %v", id, kCutLeafSets(cuts), model[id])
			}
			if err := VerifyKCuts(cuts, 4, true); err != nil {
				t.Fatal(err)
			}
		}
	})
}

func independentKCutSets(node int, left, right [][]int, limit, keep int) [][]int {
	all := []KCut{{Leaves: []int{node}, Signature: kCutSignature([]int{node})}}
	for _, a := range left {
		for _, b := range right {
			leaves, ok := independentCutUnion(a, b, limit)
			if !ok || len(leaves) == 1 {
				continue
			}
			candidate := KCut{Leaves: leaves, Signature: kCutSignature(leaves)}
			dominated := false
			for _, old := range all {
				if independentCutSubset(old.Leaves, candidate.Leaves) {
					dominated = true
					break
				}
			}
			if dominated {
				continue
			}
			kept := all[:0]
			for _, old := range all {
				if !independentCutSubset(candidate.Leaves, old.Leaves) {
					kept = append(kept, old)
				}
			}
			all = append(kept, candidate)
			if len(all) >= keep {
				goto done
			}
		}
	}
done:
	sort.SliceStable(all, func(i, j int) bool { return kCutLess(all[i], all[j]) })
	return kCutLeafSets(all)
}

func independentCutUnion(a, b []int, limit int) ([]int, bool) {
	seen := make(map[int]bool, len(a)+len(b))
	for _, v := range a {
		seen[v] = true
	}
	for _, v := range b {
		seen[v] = true
	}
	if len(seen) > limit {
		return nil, false
	}
	res := make([]int, 0, len(seen))
	for v := range seen {
		res = append(res, v)
	}
	sort.Ints(res)
	return res, true
}

func independentCutSubset(a, b []int) bool {
	seen := make(map[int]bool, len(b))
	for _, v := range b {
		seen[v] = true
	}
	for _, v := range a {
		if !seen[v] {
			return false
		}
	}
	return true
}

func kCutLeafSets(cuts []KCut) [][]int {
	sets := make([][]int, len(cuts))
	for i := range cuts {
		sets[i] = append([]int(nil), cuts[i].Leaves...)
	}
	return sets
}

func truthFromFunction(vars int, function func(int) bool) []uint64 {
	truth := make([]uint64, kCutTruthWords(vars))
	for assignment := 0; assignment < 1<<uint(vars); assignment++ {
		if function(assignment) {
			kCutSetTruthBit(truth, assignment)
		}
	}
	return truth
}

func evalKCutGraph(nodes []kCutTestNode, root int, leaves []int, assignment int) bool {
	positions := make(map[int]int, len(leaves))
	for i, leaf := range leaves {
		positions[leaf] = i
	}
	var eval func(int) bool
	eval = func(node int) bool {
		if pos, ok := positions[node]; ok {
			return assignment&(1<<uint(pos)) != 0
		}
		n := nodes[node]
		a, b := eval(n.left), eval(n.right)
		if n.invLeft {
			a = !a
		}
		if n.invRight {
			b = !b
		}
		return a && b
	}
	return eval(root)
}
