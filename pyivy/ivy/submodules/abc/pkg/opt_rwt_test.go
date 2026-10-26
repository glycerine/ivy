package abc

// Stand-alone validation for src/opt/rwt without a native oracle.

import (
	"crypto/sha256"
	"errors"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

func rwtTestTruth(graph *DecoGraph) (uint16, error) {
	var truth uint16
	for assignment := 0; assignment < 16; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		value, err := graph.Evaluate(inputs)
		if err != nil {
			return 0, err
		}
		if value {
			truth |= 1 << uint(assignment)
		}
	}
	return truth, nil
}

func rwtTestNodeTruth(nodes []RwtNode, link RwtLink) uint16 {
	truth := nodes[link.Node].Truth
	if link.Complement {
		return ^truth
	}
	return truth
}

func rwtTestVolume(nodes []RwtNode, first, second int) int {
	seen := make([]bool, len(nodes))
	result := 0
	var visit func(int)
	visit = func(id int) {
		n := nodes[id]
		if n.Used || seen[id] {
			return
		}
		seen[id] = true
		result++
		if n.XOR {
			result++
		}
		visit(n.Fan0.Node)
		visit(n.Fan1.Node)
	}
	visit(first)
	visit(second)
	return result
}

func TestRwtStaticForestAndDecompositions(t *testing.T) {
	entries, err := decodeRwtStatic()
	if err != nil {
		t.Fatal(err)
	}
	encoded := make([]byte, 2*len(entries))
	for i, entry := range entries {
		encoded[2*i] = byte(entry)
		encoded[2*i+1] = byte(entry >> 8)
	}
	if got := sha256.Sum256(encoded); got != [32]byte{0xec, 0x89, 0x0a, 0xa4, 0x4e, 0xc9, 0x3d, 0x66, 0x2d, 0x45, 0x2b, 0x34, 0x53, 0xe2, 0x6f, 0x55, 0x4c, 0x32, 0xad, 0x0b, 0xa8, 0x4f, 0x6e, 0xaa, 0x60, 0x44, 0xda, 0x99, 0xd8, 0xf1, 0x93, 0x05} {
		t.Fatalf("static forest digest=%x", got)
	}
	if len(entries) != 2*(1780+1) || entries[len(entries)-2] != 0 || entries[len(entries)-1] != 0 {
		t.Fatalf("static forest words=%d terminator=%v", len(entries), entries[len(entries)-2:])
	}
	manager, err := NewRwtManager(RwtOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	nodes, err := manager.Nodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1785 {
		t.Fatalf("forest nodes=%d", len(nodes))
	}
	stats, _ := manager.Stats()
	if stats.Considered != 1780 || stats.Classes != 139 || stats.Added != 1139 {
		t.Fatalf("stats=%+v", stats)
	}
	for id, node := range nodes {
		if node.ID != id {
			t.Fatalf("id[%d]=%d", id, node.ID)
		}
		if id < 5 {
			continue
		}
		e0, e1 := uint32(entries[2*(id-5)]), uint32(entries[2*(id-5)+1])
		wantXOR := e0&1 != 0
		e0 >>= 1
		wantFirst := RwtLink{Node: int(e0 >> 1), Complement: e0&1 != 0}
		wantSecond := RwtLink{Node: int(e1 >> 1), Complement: e1&1 != 0}
		if node.XOR != wantXOR || node.Fan0 != wantFirst || node.Fan1 != wantSecond {
			t.Fatalf("static pair node %d: xor=%v fanins=%+v/%+v want %v %+v/%+v", id, node.XOR, node.Fan0, node.Fan1, wantXOR, wantFirst, wantSecond)
		}
		left, right := rwtTestNodeTruth(nodes, node.Fan0), rwtTestNodeTruth(nodes, node.Fan1)
		want := left & right
		if node.XOR {
			want = left ^ right
		}
		if node.Truth != want {
			t.Fatalf("truth node %d: %04x != %04x", id, node.Truth, want)
		}
		wantLevel := 1 + max(nodes[node.Fan0.Node].Level, nodes[node.Fan1.Node].Level)
		if node.Level != wantLevel {
			t.Fatalf("level node %d: %d/%d", id, node.Level, wantLevel)
		}
	}
	classes, err := manager.Classes()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	expectedClasses := make([][]int, rwtClassCount)
	for truth := 0; truth < rwtFunctionCount; truth++ {
		for id, node := range nodes {
			if id < 5 || int(node.Truth) != truth {
				continue
			}
			canonical, _, _, class, canonicalErr := manager.Canonical(node.Truth)
			if canonicalErr != nil {
				t.Fatal(canonicalErr)
			}
			if node.Truth == canonical {
				expectedClasses[int(class)] = append(expectedClasses[int(class)], id)
			}
		}
	}
	for class, ids := range classes {
		if !slices.Equal(ids, expectedClasses[class]) {
			t.Fatalf("class %d order=%v want %v", class, ids, expectedClasses[class])
		}
		for _, id := range ids {
			count++
			graph, graphErr := manager.Decomposition(id)
			if graphErr != nil {
				t.Fatal(graphErr)
			}
			truth, truthErr := rwtTestTruth(graph)
			graph.Close()
			if truthErr != nil || truth != nodes[id].Truth {
				t.Fatalf("class %d node %d: %04x want %04x err=%v", class, id, truth, nodes[id].Truth, truthErr)
			}
			representative, _ := manager.ClassRepresentative(class)
			canonical, _, _, mapped, _ := manager.Canonical(nodes[id].Truth)
			if int(mapped) != class || representative != canonical {
				t.Fatalf("class mapping %d node %d", class, id)
			}
		}
	}
	if count != stats.Added {
		t.Fatalf("decompositions=%d added=%d", count, stats.Added)
	}
}

func TestRwtIndependentFiveThousandModelChecks(t *testing.T) {
	manager, err := NewRwtManager(RwtOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	nodes, _ := manager.Nodes()
	random := rand.New(rand.NewSource(0x727774))
	for trial := 0; trial < 5000; trial++ {
		id := 5 + random.Intn(len(nodes)-5)
		node := nodes[id]
		want := rwtTestNodeTruth(nodes, node.Fan0) & rwtTestNodeTruth(nodes, node.Fan1)
		if node.XOR {
			want = rwtTestNodeTruth(nodes, node.Fan0) ^ rwtTestNodeTruth(nodes, node.Fan1)
		}
		if node.Truth != want {
			t.Fatalf("trial %d node %d", trial, id)
		}
		first, second := random.Intn(id), random.Intn(id)
		got, volumeErr := manager.NodeVolume(first, second)
		if volumeErr != nil || got != rwtTestVolume(nodes, first, second) {
			t.Fatalf("volume trial %d: %d err=%v", trial, got, volumeErr)
		}
	}
}

func TestRwtLimitsCancellationAtomicityOwnershipAndErrors(t *testing.T) {
	if manager, err := NewRwtManager(RwtOptions{MaxNodes: 4}); manager != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit manager=%v err=%v", manager, err)
	}
	if manager, err := NewRwtManager(RwtOptions{MaxClasses: 221}); manager != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("class limit manager=%v err=%v", manager, err)
	}
	if manager, err := NewRwtManager(RwtOptions{Stop: func() bool { return true }}); manager != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel manager=%v err=%v", manager, err)
	}
	manager, err := NewRwtManager(RwtOptions{Precompute: true, MaxNodes: 6})
	if err != nil {
		t.Fatal(err)
	}
	before, _ := manager.Nodes()
	if _, err = manager.AddNode(RwtLink{Node: 99}, RwtLink{Node: 0}, false, 1, 1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	after, _ := manager.Nodes()
	if len(after) != len(before) {
		t.Fatal("failed add mutated forest")
	}
	if _, err = manager.AddNode(RwtLink{Node: 1, Complement: true}, RwtLink{Node: 2}, true, 1, 2); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if len(after) > 0 {
		after[0].Truth = 123
	}
	unchanged, _ := manager.Node(0)
	if unchanged.Truth != 0 {
		t.Fatal("Nodes returned borrowed storage")
	}
	text, _ := manager.StatsText()
	record, _ := manager.StatsRecord()
	if !strings.HasPrefix(text, "Rewriting statistics:\n") || string(record) != "0 0 0 0\n" {
		t.Fatalf("text=%q record=%q", text, record)
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Nodes(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = manager.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	traversal, err := NewRwtManager(RwtOptions{Precompute: true})
	if err != nil {
		t.Fatal(err)
	}
	traversal.traversalID = 0x8ffffffe
	traversal.nodes[0].TraversalID = 123
	if err = traversal.IncrementTraversalID(); err != nil {
		t.Fatal(err)
	}
	if traversal.traversalID != 1 || traversal.nodes[0].TraversalID != 0 {
		t.Fatalf("traversal rollover id=%x mark=%d", traversal.traversalID, traversal.nodes[0].TraversalID)
	}
	if err = traversal.Close(); err != nil {
		t.Fatal(err)
	}
	ready, err := NewRwtManager(RwtOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ready.AddVariable(0x5555, true); !errors.Is(err, ErrInvariant) {
		t.Fatalf("add variable after preprocess: %v", err)
	}
	if _, err = ready.AddNode(RwtLink{Node: 1}, RwtLink{Node: 2}, false, 1, 1); !errors.Is(err, ErrInvariant) {
		t.Fatalf("add node after preprocess: %v", err)
	}
	if err = ready.LoadStatic(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("reload after preprocess: %v", err)
	}
	if err = ready.Close(); err != nil {
		t.Fatal(err)
	}
	dirty, err := NewRwtManager(RwtOptions{Precompute: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = dirty.AddVariable(0x5555, true); err != nil {
		t.Fatal(err)
	}
	if err = dirty.LoadStatic(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("load after construction: %v", err)
	}
	if err = dirty.Close(); err != nil {
		t.Fatal(err)
	}
	limited, err := NewRwtManager(RwtOptions{Precompute: true, MaxNodes: 100})
	if err != nil {
		t.Fatal(err)
	}
	limitedNodes, _ := limited.Nodes()
	limitedStats, _ := limited.Stats()
	if err = limited.LoadStatic(); !errors.Is(err, ErrLimit) {
		t.Fatalf("limited load: %v", err)
	}
	limitedAfter, _ := limited.Nodes()
	limitedStatsAfter, _ := limited.Stats()
	if !slices.Equal(limitedNodes, limitedAfter) || limitedStats != limitedStatsAfter {
		t.Fatal("limited load was not atomic")
	}
	if err = limited.Close(); err != nil {
		t.Fatal(err)
	}
	stop := false
	atomicManager, err := NewRwtManager(RwtOptions{Precompute: true, Stop: func() bool { return stop }})
	if err != nil {
		t.Fatal(err)
	}
	beforeNodes, _ := atomicManager.Nodes()
	beforeStats, _ := atomicManager.Stats()
	stop = true
	if err = atomicManager.LoadStatic(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	afterNodes, _ := atomicManager.Nodes()
	afterStats, _ := atomicManager.Stats()
	if len(afterNodes) != len(beforeNodes) || afterStats != beforeStats {
		t.Fatal("cancelled load was not atomic")
	}
	stop = false
	if err = atomicManager.LoadStatic(); err != nil {
		t.Fatal(err)
	}
	stop = true
	loadedNodes, _ := atomicManager.Nodes()
	if err = atomicManager.Preprocess(); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel preprocess: %v", err)
	}
	afterCancelledPreprocess, _ := atomicManager.Nodes()
	if !slices.Equal(loadedNodes, afterCancelledPreprocess) {
		t.Fatal("cancelled preprocess mutated nodes")
	}
	if classes, classesErr := atomicManager.Classes(); classesErr != nil || len(classes) != 0 {
		t.Fatalf("cancelled preprocess classes=%v err=%v", classes, classesErr)
	}
	stop = false
	if err = atomicManager.Preprocess(); err != nil {
		t.Fatal(err)
	}
	if graph, graphErr := atomicManager.ResultGraph(); graphErr != nil || graph != nil {
		t.Fatalf("result graph=%v err=%v", graph, graphErr)
	}
	practical, tableErr := atomicManager.PracticalTable()
	if tableErr != nil || len(practical) != 1<<16 || !practical[0x6996] || practical[0x6997] {
		t.Fatalf("practical err=%v", tableErr)
	}
	if err = atomicManager.IncrementTraversalID(); err != nil {
		t.Fatal(err)
	}
	if err = atomicManager.Close(); err != nil {
		t.Fatal(err)
	}
	if err = RwtPrecompute(RwtOptions{}); err != nil {
		t.Fatal(err)
	}
}

func FuzzRwtTruthAndVolume(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 0x727774, 1<<63 - 1} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		manager, err := NewRwtManager(RwtOptions{Precompute: true, MaxNodes: 96})
		if err != nil {
			t.Fatal(err)
		}
		defer manager.Close()
		random := rand.New(rand.NewSource(int64(seed)))
		for trial := 0; trial < 32; trial++ {
			nodes, _ := manager.Nodes()
			a, b := random.Intn(len(nodes)), random.Intn(len(nodes))
			xor := random.Intn(4) == 0
			level := 1 + max(nodes[a].Level, nodes[b].Level)
			volume := 1 + rwtTestVolume(nodes, a, b) + rwtBoolToInt(xor)
			id, addErr := manager.AddNode(RwtLink{Node: a}, RwtLink{Node: b}, xor, level, volume)
			if addErr != nil {
				t.Fatal(addErr)
			}
			gotNode, _ := manager.Node(id)
			want := nodes[a].Truth & nodes[b].Truth
			if xor {
				want = nodes[a].Truth ^ nodes[b].Truth
			}
			if gotNode.Truth != want {
				t.Fatalf("node %d", id)
			}
			nodes, _ = manager.Nodes()
			a, b = random.Intn(len(nodes)), random.Intn(len(nodes))
			got, e := manager.NodeVolume(a, b)
			if e != nil || got != rwtTestVolume(nodes, a, b) {
				t.Fatalf("volume %d %d", a, b)
			}
		}
	})
}
