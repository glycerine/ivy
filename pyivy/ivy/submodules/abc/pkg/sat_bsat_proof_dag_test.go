package abc

// Stand-alone tests for the satProof packed proof-DAG replacement.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func satProofDAGClause(t testing.TB, clause int, pivotA bool) SatProofDAGRef {
	t.Helper()
	reference, err := MakeSatProofDAGClauseRef(clause, pivotA)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}

func satProofDAGNode(t testing.TB, node int, pivotA bool) SatProofDAGRef {
	t.Helper()
	reference, err := MakeSatProofDAGNodeRef(node, pivotA)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}

func satProofDAGAdd(t testing.TB, dag *SatProofDAG, references ...SatProofDAGRef) int {
	t.Helper()
	handle, err := dag.AddNode(SatProofDAGNode{Antecedents: references})
	if err != nil {
		t.Fatal(err)
	}
	return handle
}

func TestSatProofDAGCollectionCoreAndOwnership(t *testing.T) {
	dag, err := NewSatProofDAG(SatProofDAGLimits{})
	if err != nil {
		t.Fatal(err)
	}
	n0 := satProofDAGAdd(t, dag, satProofDAGClause(t, 3, false), satProofDAGClause(t, 1, true))
	n1 := satProofDAGAdd(t, dag, satProofDAGNode(t, n0, false), satProofDAGClause(t, 2, false))
	n2 := satProofDAGAdd(t, dag, satProofDAGClause(t, 4, true), satProofDAGNode(t, n1, false))
	postorder, err := dag.CollectUsedPostorder([]int{n2})
	if err != nil || !reflect.DeepEqual(postorder, []int{0, 1, 2}) {
		t.Fatalf("postorder = %v, err=%v", postorder, err)
	}
	sorted, err := dag.CollectUsedSorted([]int{n2, -1, n1})
	if err != nil || !reflect.DeepEqual(sorted, []int{0, 1, 2}) {
		t.Fatalf("sorted = %v, err=%v", sorted, err)
	}
	if count, err := dag.MarkUsedCount([]int{n2}); err != nil || count != 3 {
		t.Fatalf("count = %d, err=%v", count, err)
	}
	core, err := dag.CoreFromUsed(sorted)
	if err != nil || !reflect.DeepEqual(core, []int{3, 1, 2, 4}) {
		t.Fatalf("core = %v, err=%v", core, err)
	}
	derived, err := dag.DeriveCore(n2)
	if err != nil || !reflect.DeepEqual(derived, []int{1, 2, 3, 4}) {
		t.Fatalf("derived core = %v, err=%v", derived, err)
	}
	node, err := dag.Node(n0)
	if err != nil {
		t.Fatal(err)
	}
	node.Antecedents[0].Clause = 99
	again, _ := dag.Node(n0)
	if again.Antecedents[0].Clause != 3 {
		t.Fatal("node snapshot aliases DAG")
	}
	if err = dag.SetLabel(n1, 17); err != nil {
		t.Fatal(err)
	}
	if err = dag.ClearLabels([]int{n1}); err != nil {
		t.Fatal(err)
	}
	again, _ = dag.Node(n1)
	if again.Label != 0 {
		t.Fatalf("label = %d", again.Label)
	}
	if err = dag.Check(); err != nil {
		t.Fatal(err)
	}
	if stats := dag.Stats(); stats.Nodes != 3 || stats.Antecedents != 6 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestSatProofDAGReductionAndResize(t *testing.T) {
	dag, _ := NewSatProofDAG(SatProofDAGLimits{})
	unused := satProofDAGAdd(t, dag, satProofDAGClause(t, 9, false))
	used := satProofDAGAdd(t, dag, satProofDAGClause(t, 1, false), satProofDAGClause(t, 2, true), satProofDAGClause(t, 8, false))
	if err := dag.ResizeAntecedents(used, 2); err != nil {
		t.Fatal(err)
	}
	root := satProofDAGAdd(t, dag, satProofDAGNode(t, used, false), satProofDAGClause(t, 3, false))
	if err := dag.SetLabel(root, 44); err != nil {
		t.Fatal(err)
	}
	reduction, err := dag.Reduce([]int{-1, root}, used)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reduction.Roots, []int{-1, 1}) || !reflect.DeepEqual(reduction.OldToNew, []int{-1, 0, 1}) || reduction.Pivot != 0 {
		t.Fatalf("reduction = %+v", reduction)
	}
	if stats := dag.Stats(); stats.Nodes != 2 || stats.Antecedents != 4 {
		t.Fatalf("stats = %+v", stats)
	}
	if core, err := dag.DeriveCore(1); err != nil || !reflect.DeepEqual(core, []int{1, 2, 3}) {
		t.Fatalf("core = %v, err=%v", core, err)
	}
	rootNode, _ := dag.Node(1)
	if rootNode.Label != 0 || rootNode.Antecedents[0].Node != 0 {
		t.Fatalf("remapped root = %+v", rootNode)
	}
	if err := dag.Check(); err != nil {
		t.Fatal(err)
	}
	_ = unused
}

func TestSatProofDAGResolutionVerification(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	y := satStoreTestLit(t, 1, false)
	clauses := [][]SatLit{{x, y}, {x.Not()}, {y.Not()}}
	dag, _ := NewSatProofDAG(SatProofDAGLimits{})
	yClause := satProofDAGAdd(t, dag, satProofDAGClause(t, 0, false), satProofDAGClause(t, 1, true))
	empty := satProofDAGAdd(t, dag, satProofDAGNode(t, yClause, false), satProofDAGClause(t, 2, false))
	results, err := dag.Verify([]int{yClause, empty}, clauses)
	if err != nil || !reflect.DeepEqual(results, [][]SatLit{{y}, {}}) {
		t.Fatalf("results = %v, err=%v", results, err)
	}
	if core, err := dag.DeriveCore(empty); err != nil || !reflect.DeepEqual(core, []int{0, 1, 2}) {
		t.Fatalf("core = %v, err=%v", core, err)
	}
	bad, _ := NewSatProofDAG(SatProofDAGLimits{})
	badRoot := satProofDAGAdd(t, bad, satProofDAGClause(t, 0, false), satProofDAGClause(t, 0, false))
	if _, err := bad.Verify([]int{badRoot}, clauses); err == nil {
		t.Fatal("invalid resolution verified")
	}
}

func TestSatProofDAGDeepIterativeTraversal(t *testing.T) {
	const depth = 10000
	dag, _ := NewSatProofDAG(SatProofDAGLimits{MaxNodes: depth, MaxAntecedents: depth})
	previous := satProofDAGAdd(t, dag, satProofDAGClause(t, 7, false))
	for i := 1; i < depth; i++ {
		previous = satProofDAGAdd(t, dag, satProofDAGNode(t, previous, i&1 != 0))
	}
	used, err := dag.CollectUsedPostorder([]int{previous})
	if err != nil || len(used) != depth || used[0] != 0 || used[len(used)-1] != depth-1 {
		t.Fatalf("used len=%d endpoints=%d/%d err=%v", len(used), used[0], used[len(used)-1], err)
	}
	core, err := dag.DeriveCore(previous)
	if err != nil || !reflect.DeepEqual(core, []int{7}) {
		t.Fatalf("core = %v, err=%v", core, err)
	}
}

func TestSatProofDAGErrorsLimitsCancellationAndLifecycle(t *testing.T) {
	if _, err := NewSatProofDAG(SatProofDAGLimits{MaxNodes: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative limit: %v", err)
	}
	if _, err := SatProofDAGRecordWords(0); !errors.Is(err, ErrDimension) {
		t.Fatalf("zero words: %v", err)
	}
	for _, row := range []struct {
		antecedents int
		words       int
	}{
		{antecedents: 1, words: 2},
		{antecedents: 2, words: 2},
		{antecedents: 3, words: 3},
		{antecedents: 4, words: 3},
		{antecedents: 5, words: 4},
	} {
		if got, err := SatProofDAGRecordWords(row.antecedents); err != nil || got != row.words {
			t.Fatalf("words(%d) = %d, %v; want %d", row.antecedents, got, err, row.words)
		}
	}
	if _, err := MakeSatProofDAGNodeRef(-1, false); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad node ref: %v", err)
	}
	if _, err := MakeSatProofDAGClauseRef(-1, false); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad clause ref: %v", err)
	}
	dag, _ := NewSatProofDAG(SatProofDAGLimits{MaxNodes: 1, MaxAntecedents: 1, MaxClauseID: 2})
	if _, err := dag.AddNode(SatProofDAGNode{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("empty node: %v", err)
	}
	root := satProofDAGAdd(t, dag, satProofDAGClause(t, 2, false))
	if _, err := dag.AddNode(SatProofDAGNode{Antecedents: []SatProofDAGRef{satProofDAGNode(t, root, false)}}); !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit: %v", err)
	}
	if err := dag.ResizeAntecedents(root, 0); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad resize: %v", err)
	}
	if _, err := dag.CollectUsedSorted([]int{9}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("bad root: %v", err)
	}
	cancelled, _ := NewSatProofDAG(SatProofDAGLimits{Stop: func() bool { return true }})
	cancelRoot := satProofDAGAdd(t, cancelled, satProofDAGClause(t, 0, false))
	if _, err := cancelled.CollectUsedSorted([]int{cancelRoot}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %v", err)
	}
	dag.Stop()
	dag.Stop()
	if !dag.Stats().Stopped {
		t.Fatal("DAG did not stop")
	}
	if _, err := dag.Node(0); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("use after stop: %v", err)
	}
}

func FuzzSatProofDAGCoreAndReduction(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2, 3, 4}, {255, 17, 33, 65, 129}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		count := 1 + len(data)%128
		dag, err := NewSatProofDAG(SatProofDAGLimits{MaxNodes: count, MaxAntecedents: 3 * count, MaxClauseID: 255})
		if err != nil {
			t.Fatal(err)
		}
		model := make([][]SatProofDAGRef, count)
		for node := 0; node < count; node++ {
			value := byte(node)
			if len(data) > 0 {
				value = data[node%len(data)]
			}
			references := []SatProofDAGRef{satProofDAGClause(t, int(value), value&1 != 0)}
			if node > 0 && value&2 != 0 {
				parent := int(value) % node
				references = append(references, satProofDAGNode(t, parent, value&4 != 0))
			}
			if value&8 != 0 {
				references = append(references, satProofDAGClause(t, int(value^0x55), value&16 != 0))
			}
			model[node] = slices.Clone(references)
			if handle := satProofDAGAdd(t, dag, references...); handle != node {
				t.Fatalf("handle = %d, want %d", handle, node)
			}
		}
		root := count - 1
		visited := make([]bool, count)
		clauseSeen := make([]bool, 256)
		var visit func(int)
		visit = func(node int) {
			if visited[node] {
				return
			}
			visited[node] = true
			for _, reference := range model[node] {
				if reference.Node >= 0 {
					visit(reference.Node)
				} else {
					clauseSeen[reference.Clause] = true
				}
			}
		}
		visit(root)
		want := make([]int, 0)
		for clause, seen := range clauseSeen {
			if seen {
				want = append(want, clause)
			}
		}
		core, err := dag.DeriveCore(root)
		if err != nil || !reflect.DeepEqual(core, want) {
			t.Fatalf("core = %v, err=%v; want %v", core, err, want)
		}
		reduction, err := dag.Reduce([]int{root}, root)
		if err != nil || reduction.Pivot != len(dag.nodes)-1 || reduction.Roots[0] != len(dag.nodes)-1 {
			t.Fatalf("reduction = %+v, err=%v", reduction, err)
		}
		if err = dag.Check(); err != nil {
			t.Fatal(err)
		}
		after, err := dag.DeriveCore(reduction.Roots[0])
		if err != nil || !reflect.DeepEqual(after, want) {
			t.Fatalf("reduced core = %v, err=%v; want %v", after, err, want)
		}
	})
}
