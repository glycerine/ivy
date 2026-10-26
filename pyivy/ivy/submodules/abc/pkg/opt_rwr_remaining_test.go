package abc

// Stand-alone validation for src/opt/rwr excluding the separately ported
// rwrTemp.c workflow in opt_rwr_temp_test.go.

import (
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

func rwrModelTruth(nodes []RwrNode, l RwtLink) uint16 {
	v := nodes[l.Node].Truth
	if l.Complement {
		return ^v
	}
	return v
}
func rwrModelVolume(nodes []RwrNode, a, b int, xorCost bool) int {
	seen := make([]bool, len(nodes))
	n := 0
	var visit func(int)
	visit = func(id int) {
		v := nodes[id]
		if v.Used || seen[id] {
			return
		}
		seen[id] = true
		n++
		if xorCost && v.XOR {
			n++
		}
		visit(v.Fan0.Node)
		visit(v.Fan1.Node)
	}
	visit(a)
	visit(b)
	return n
}
func rwrGraphTruth(g *DecoGraph) (uint16, error) {
	var truth uint16
	for a := 0; a < 16; a++ {
		v, e := g.Evaluate([]bool{a&1 != 0, a&2 != 0, a&4 != 0, a&8 != 0})
		if e != nil {
			return 0, e
		}
		if v {
			truth |= 1 << uint(a)
		}
	}
	return truth, nil
}

func TestRwrStaticLibraryExact(t *testing.T) {
	m, e := NewRwrManager(RwrOptions{})
	if e != nil {
		t.Fatal(e)
	}
	defer m.Close()
	nodes, e := m.Nodes()
	if e != nil {
		t.Fatal(e)
	}
	if len(nodes) != 1785 {
		t.Fatalf("nodes=%d", len(nodes))
	}
	s, _ := m.Stats()
	if s.Considered != 1780 || s.Added != 1139 || s.Classes != 139 {
		t.Fatalf("stats=%+v", s)
	}
	words, e := m.StaticArray()
	if e != nil {
		t.Fatal(e)
	}
	want, e := decodeRwtStatic()
	if e != nil || !slices.Equal(words, want) {
		t.Fatalf("static words=%d err=%v", len(words), e)
	}
	practical, e := m.PracticalTable()
	if e != nil || len(practical) != 1<<16 || !practical[0x6996] || practical[0x6997] {
		t.Fatalf("practical err=%v", e)
	}
	binaryData, e := m.BinaryLibrary()
	if e != nil {
		t.Fatal(e)
	}
	decoded, e := DecodeRwrBinaryLibrary(binaryData, 1780)
	if e != nil || !slices.Equal(decoded, want) {
		t.Fatalf("binary err=%v", e)
	}
	arrayText, e := m.StaticArrayText()
	if e != nil || !strings.HasPrefix(string(arrayText), "static unsigned short s_RwtAigSubgraphs[] = \n{") {
		t.Fatalf("array text err=%v", e)
	}
	classes, _ := m.Classes()
	expectedClasses := make([][]int, len(classes))
	for truth, ids := range m.core.table {
		class := int(m.core.classMap[truth])
		expectedClasses[class] = append(expectedClasses[class], ids...)
	}
	for class := range classes {
		if !slices.Equal(classes[class], expectedClasses[class]) {
			t.Fatalf("class %d is not in source truth/table order", class)
		}
	}
	retained := 0
	for id, n := range nodes {
		if n.ID != id {
			t.Fatalf("id %d/%d", n.ID, id)
		}
		if id < 5 {
			continue
		}
		left, right := rwrModelTruth(nodes, n.Fan0), rwrModelTruth(nodes, n.Fan1)
		truth := left & right
		if n.XOR {
			truth = left ^ right
		}
		if n.Truth != truth {
			t.Fatalf("truth id=%d", id)
		}
		if n.Level != 1+max(nodes[n.Fan0.Node].Level, nodes[n.Fan1.Node].Level) {
			t.Fatalf("level id=%d", id)
		}
	}
	for _, ids := range classes {
		for _, id := range ids {
			retained++
			g, e := m.Decomposition(id)
			if e != nil {
				t.Fatal(e)
			}
			truth, e := rwrGraphTruth(g)
			g.Close()
			if e != nil || truth != nodes[id].Truth {
				t.Fatalf("decomposition id=%d got=%04x want=%04x err=%v", id, truth, nodes[id].Truth, e)
			}
		}
	}
	if retained != 1139 {
		t.Fatalf("retained=%d", retained)
	}
	if line, _ := m.NodeText(0); !strings.Contains(line, "Const1") {
		t.Fatalf("line=%q", line)
	}
	if line, _ := m.NodeText(5); !strings.Contains(line, "a") {
		t.Fatalf("line=%q", line)
	}
	text, e := m.LibraryText()
	if e != nil || !strings.Contains(text, "Class") || !strings.Contains(text, "Roots =") {
		t.Fatalf("library err=%v", e)
	}
}

func TestRwrFiveThousandIndependentModels(t *testing.T) {
	m, e := NewRwrManager(RwrOptions{})
	if e != nil {
		t.Fatal(e)
	}
	defer m.Close()
	nodes, _ := m.Nodes()
	r := rand.New(rand.NewSource(0x727772))
	for trial := 0; trial < 5000; trial++ {
		id := 5 + r.Intn(len(nodes)-5)
		n := nodes[id]
		truth := rwrModelTruth(nodes, n.Fan0) & rwrModelTruth(nodes, n.Fan1)
		if n.XOR {
			truth = rwrModelTruth(nodes, n.Fan0) ^ rwrModelTruth(nodes, n.Fan1)
		}
		if truth != n.Truth {
			t.Fatalf("trial=%d id=%d", trial, id)
		}
		a, b := r.Intn(id), r.Intn(id)
		got, e := m.NodeVolume(a, b)
		if e != nil || got != rwrModelVolume(nodes, a, b, true) {
			t.Fatalf("volume trial=%d got=%d err=%v", trial, got, e)
		}
		entry := int(nodes[id].Truth)
		volume, roots, sum, e := m.BushMetrics(entry)
		if e != nil || volume < 0 || roots < 0 || sum < volume {
			t.Fatalf("bush trial=%d %d %d %d %v", trial, volume, roots, sum, e)
		}
	}
}

func TestRwrTryNodeLimitsAtomicityOwnershipAndExplorers(t *testing.T) {
	if m, e := NewRwrManager(RwrOptions{MaxNodes: 4}); m != nil || !errors.Is(e, ErrLimit) {
		t.Fatalf("manager=%v err=%v", m, e)
	}
	if m, e := NewRwrManager(RwrOptions{Stop: func() bool { return true }}); m != nil || !errors.Is(e, ErrCancelled) {
		t.Fatalf("manager=%v err=%v", m, e)
	}
	m, e := NewRwrManager(RwrOptions{})
	if e != nil {
		t.Fatal(e)
	}
	before, _ := m.Nodes()
	if _, _, e = m.TryNode(RwtLink{Node: 9999}, RwtLink{Node: 1}, false, 1, 1); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	after, _ := m.Nodes()
	if len(after) != len(before) {
		t.Fatal("failed TryNode mutated")
	}
	bad := make([]byte, 12)
	binary.LittleEndian.PutUint32(bad, 1)
	binary.LittleEndian.PutUint32(bad[4:], uint32(0xffff))
	binary.LittleEndian.PutUint32(bad[8:], uint32(2))
	if e = m.LoadBinaryLibrary(bad, 1); !errors.Is(e, ErrFormat) {
		t.Fatalf("bad load err=%v", e)
	}
	afterBad, _ := m.Nodes()
	if len(afterBad) != len(before) {
		t.Fatal("bad binary load mutated manager")
	}
	after[0].Truth = 7
	again, _ := m.Nodes()
	if again[0].Truth != 0 {
		t.Fatal("borrowed nodes")
	}
	if _, _, e = m.TryNode(RwtLink{Node: 1, Complement: true}, RwtLink{Node: 2}, true, 1, 2); !errors.Is(e, ErrInvariant) {
		t.Fatal(e)
	}
	if e = m.ScoreClean(); e != nil {
		t.Fatal(e)
	}
	report, e := m.ScoreReport()
	if e != nil || len(report) != 0 {
		t.Fatalf("report=%v err=%v", report, e)
	}
	record, _ := m.StatsRecord()
	if string(record) != "0 0 0 0\n" {
		t.Fatalf("record=%q", record)
	}
	if e = m.Close(); e != nil {
		t.Fatal(e)
	}
	if e = m.Close(); !errors.Is(e, ErrInvariant) {
		t.Fatal(e)
	}
	x, e := NewRwrExplorer4()
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range []uint16{1, 0xfffe, 0x6996, 0x6996} {
		if e = x.Count(v); e != nil {
			t.Fatal(e)
		}
	}
	xr, e := x.Report()
	if e != nil || xr.Cuts != 4 || xr.Classes != 2 {
		t.Fatalf("report=%+v err=%v", xr, e)
	}
	if e = x.Close(); e != nil {
		t.Fatal(e)
	}
	y := NewRwrExplorer5()
	for _, v := range []uint32{1, 1, 3, 0x69969669} {
		if e = y.Count(v); e != nil {
			t.Fatal(e)
		}
	}
	yr, e := y.Report()
	if e != nil || yr.Cuts != 4 || len(yr.NN) != 3 || len(yr.NPN) == 0 {
		t.Fatalf("report=%+v err=%v", yr, e)
	}
	if e = y.Close(); e != nil {
		t.Fatal(e)
	}
	calls := 0
	if p, e := NewRwrManager(RwrOptions{Precompute: true, MaxOperations: 1000, Stop: func() bool { calls++; return calls > 20 }}); p != nil || !errors.Is(e, ErrCancelled) {
		t.Fatalf("precompute cancel manager=%v err=%v calls=%d", p, e, calls)
	}
}

func TestRwrBinaryLoadStateTraversalAndEncodingLimits(t *testing.T) {
	source, err := NewRwrManager(RwrOptions{})
	if err != nil {
		t.Fatal(err)
	}
	data, err := source.BinaryLibrary()
	if err != nil {
		t.Fatal(err)
	}
	if err = source.LoadBinaryLibrary(data, 0); !errors.Is(err, ErrInvariant) {
		t.Fatalf("preprocessed load error = %v", err)
	}
	_ = source.Close()

	loaded, err := NewRwrManagerFromBinary(data, 0, RwrOptions{MaxNodes: 2000})
	if err != nil {
		t.Fatal(err)
	}
	nodes, _ := loaded.Nodes()
	if len(nodes) != 1785 {
		t.Fatal(len(nodes))
	}
	_ = loaded.Close()

	broken := make([]byte, 20)
	binary.LittleEndian.PutUint32(broken, 2)
	binary.LittleEndian.PutUint32(broken[4:], 8)
	binary.LittleEndian.PutUint32(broken[8:], 2)
	binary.LittleEndian.PutUint32(broken[12:], 0xffff)
	binary.LittleEndian.PutUint32(broken[16:], 2)
	atomic, err := newRwrEmptyManager(RwrOptions{MaxNodes: 32})
	if err != nil {
		t.Fatal(err)
	}
	beforeTraversal := atomic.core.traversalID
	if err = atomic.LoadBinaryLibrary(broken, 2); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	if atomic.core.traversalID != beforeTraversal || len(atomic.core.nodes) != 5 {
		t.Fatal("failed load changed manager")
	}
	atomic.nodes = append(atomic.nodes, RwrNode{RwtNode: RwtNode{Fan0: RwtLink{Node: math.MaxUint16>>2 + 1}, Fan1: RwtLink{Node: 1}}})
	if _, err = atomic.StaticArray(); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	_ = atomic.Close()

	stop := false
	cancelled, err := newRwrEmptyManager(RwrOptions{MaxNodes: 64, Stop: func() bool { return stop }})
	if err != nil {
		t.Fatal(err)
	}
	beforeTraversal = cancelled.core.traversalID
	stop = true
	if err = cancelled.Precompute(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if cancelled.core.traversalID != beforeTraversal || len(cancelled.core.nodes) != 5 {
		t.Fatal("cancelled precompute changed manager")
	}
	_ = cancelled.Close()
}

func TestRwrCutUnionTraversalAndFourInputContract(t *testing.T) {
	g := NewAigManager()
	leaves := make([]AigLit, 4)
	for i := range leaves {
		leaves[i], _ = g.CreateCI()
	}
	n1, _ := g.And(leaves[0], leaves[1])
	n2, _ := g.And(n1, leaves[2])
	root, _ := g.And(n2, leaves[3])
	cuts := []AigCNFCut4{
		{Root: root.ID(), Leaves: []AigID{n1.ID(), leaves[2].ID(), leaves[3].ID()}},
		{Root: root.ID(), Leaves: []AigID{leaves[0].ID(), leaves[1].ID(), leaves[2].ID(), leaves[3].ID()}},
	}
	count, err := RwrCutNodeCount(g, root.ID(), cuts)
	if err != nil || count != 7 {
		t.Fatalf("union count = %d, %v; want 7", count, err)
	}
	manager, _ := NewRwrManager(RwrOptions{})
	wide := AigCNFCut4{Root: root.ID(), Leaves: []AigID{leaves[0].ID(), leaves[1].ID(), leaves[2].ID(), leaves[3].ID(), n1.ID()}}
	if _, err = manager.RewriteNode(g, root.ID(), []AigCNFCut4{wide}, RwrRewriteOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	_ = manager.Close()
}

func TestRwrRewriteEvaluationAndHelpers(t *testing.T) {
	g := NewAigManager()
	leaf := make([]AigLit, 4)
	for i := range leaf {
		leaf[i], _ = g.CreateCI()
	}
	left, _ := g.And(leaf[0], leaf[1])
	right, _ := g.And(leaf[2], leaf[3])
	root, _ := g.And(left, right)
	_, _ = g.CreateCO(leaf[0])
	_, _ = g.CreateCO(leaf[1])
	truthResult, e := g.CutTruth(root, []AigID{leaf[0].ID(), leaf[1].ID(), leaf[2].ID(), leaf[3].ID()}, AigCutTruthOptions{})
	if e != nil {
		t.Fatal(e)
	}
	words := truthResult.Table.Words()
	truth := uint16(words[0])
	cut := AigCNFCut4{Root: root.ID(), Leaves: []AigID{leaf[0].ID(), leaf[1].ID(), leaf[2].ID(), leaf[3].ID()}, Truth: truth}
	m, e := NewRwrManager(RwrOptions{})
	if e != nil {
		t.Fatal(e)
	}
	defer m.Close()
	result, e := m.RewriteNode(g, root.ID(), []AigCNFCut4{cut}, RwrRewriteOptions{UseZeroGain: true})
	if e != nil {
		t.Fatal(e)
	}
	if result.Gain < 0 || result.Graph == nil || len(result.Leaves) != 4 {
		t.Fatalf("result=%+v", result)
	}
	leafIDs := make([]AigID, len(result.Leaves))
	for i, lit := range result.Leaves {
		leafIDs[i] = lit.ID()
	}
	mffc, e := g.MFFCLabelCut(root.ID(), leafIDs, AigMFFCOptions{MaxVisits: 1 << 20})
	if e != nil {
		t.Fatal(e)
	}
	class := int(m.core.classMap[cut.Truth])
	wantID, wantGain := -1, -1
	for _, id := range m.core.classes[class] {
		candidate := m.core.graphs[id]
		if candidate == nil {
			continue
		}
		added, err := rwrNetworkCount(g, candidate, result.Leaves, mffc.Members, mffc.Size, math.MaxInt)
		if err != nil {
			t.Fatal(err)
		}
		gain := mffc.Size - added
		if added >= 0 && gain > wantGain {
			wantID, wantGain = id, gain
		}
	}
	if result.LibraryNode != wantID || result.Gain != wantGain {
		t.Fatalf("strict source-order winner=(%d,%d), want=(%d,%d)", result.LibraryNode, result.Gain, wantID, wantGain)
	}
	result.Graph.Close()
	stored, leaves, complement, e := m.Result()
	if e != nil || stored == nil || len(leaves) != 4 || complement != result.Complement {
		t.Fatalf("stored err=%v", e)
	}
	stored.Close()
	beforeStats, _ := m.Stats()
	beforeNodes, _ := m.Nodes()
	wide := cut
	wide.Leaves = append(slices.Clone(cut.Leaves), root.ID())
	if _, e = m.RewriteNode(g, root.ID(), []AigCNFCut4{cut, wide}, RwrRewriteOptions{UseZeroGain: true}); !errors.Is(e, ErrDimension) {
		t.Fatalf("transactional rewrite error=%v", e)
	}
	afterStats, _ := m.Stats()
	afterNodes, _ := m.Nodes()
	if beforeStats != afterStats || !slices.Equal(beforeNodes, afterNodes) {
		t.Fatal("failed multi-cut rewrite changed manager state")
	}
	preserved, preservedLeaves, preservedComplement, e := m.Result()
	if e != nil || preserved == nil || !slices.Equal(preservedLeaves, leaves) || preservedComplement != complement {
		t.Fatalf("failed multi-cut rewrite replaced result: err=%v", e)
	}
	preserved.Close()
	boolean, e := RwrCutIsBoolean(g, root.ID(), cut.Leaves)
	if e != nil || boolean {
		t.Fatalf("boolean=%v err=%v", boolean, e)
	}
	count, e := RwrCutNodeCount(g, root.ID(), []AigCNFCut4{cut})
	if e != nil || count < 3 {
		t.Fatalf("count=%d err=%v", count, e)
	}
	depth, e := RwrNodeDepth(g, root.ID(), cut.Leaves)
	if e != nil || depth != 2 {
		t.Fatalf("depth=%d err=%v", depth, e)
	}
}

func FuzzRwrBinaryLibrary(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 0x727772, 1<<63 - 1} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		r := rand.New(rand.NewSource(int64(seed)))
		n := r.Intn(65)
		data := make([]byte, 4+8*n)
		binary.LittleEndian.PutUint32(data, uint32(n))
		want := make([]uint16, 0, 2*n+2)
		for i := 0; i < 2*n; i++ {
			v := uint16(r.Uint32())
			binary.LittleEndian.PutUint32(data[4+4*i:], uint32(v))
			want = append(want, v)
		}
		want = append(want, 0, 0)
		got, e := DecodeRwrBinaryLibrary(data, 64)
		if e != nil {
			t.Fatal(e)
		}
		if !slices.Equal(got, want) {
			t.Fatal("round trip")
		}
		if n > 0 {
			truncated := data[:len(data)-1]
			if _, e = DecodeRwrBinaryLibrary(truncated, 64); !errors.Is(e, ErrFormat) {
				t.Fatalf("truncated err=%v", e)
			}
		}
	})
}
