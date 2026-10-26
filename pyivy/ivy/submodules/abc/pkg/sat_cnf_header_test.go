package abc

// Stand-alone tests for the public CNF header replacements.

import (
	"errors"
	"slices"
	"testing"
)

func headerTruth(t testing.TB, variables int, word uint64) *TruthTable {
	t.Helper()
	truth, err := TruthTableFromWords(variables, []uint64{word}, 0)
	if err != nil {
		t.Fatal(err)
	}
	return truth
}

func TestCNFHeaderCutAccessAndSOPCost(t *testing.T) {
	cut := AigCNFCut{Root: 7, Leaves: []AigID{1, 3}, Truth: headerTruth(t, 2, 0x8888888888888888)}
	if CNFCutLeafCount(cut) != 2 || !slices.Equal(CNFCutLeaves(cut), []AigID{1, 3}) || !CNFCutTruth(cut).Equal(cut.Truth) {
		t.Fatal("cut accessors")
	}
	leaves := CNFCutLeaves(cut)
	leaves[0] = 6
	truth := CNFCutTruth(cut)
	if err := truth.Set(0, true); err != nil || cut.Leaves[0] != 1 {
		t.Fatal("owned accessors", err)
	}
	for _, row := range []struct {
		truth uint16
		cost  int
	}{{0, 1}, {0xffff, 1}, {0xaaaa, 2}, {0x8888, 3}, {0x6996, 16}} {
		got, err := CNFCutSOPCost4(AigCNFCut4{Truth: row.truth})
		if err != nil || got != row.cost {
			t.Fatalf("cost %#x: %d %v", row.truth, got, err)
		}
	}
}

func TestCNFHeaderBestCuts(t *testing.T) {
	table, err := NewCNFBestCutTable(8, CNFHeaderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cut := AigCNFCut{Root: 7, Leaves: []AigID{1, 3}, Truth: headerTruth(t, 2, 0xeeeeeeeeeeeeeeee)}
	if err = table.SetBestCut(7, &cut); err != nil {
		t.Fatal(err)
	}
	cut.Leaves[0] = 2
	got, found, err := table.BestCut(7)
	if err != nil || !found || !slices.Equal(got.Leaves, []AigID{1, 3}) {
		t.Fatal(got, found, err)
	}
	got.Leaves[0] = 5
	again, found, err := table.BestCut(7)
	if err != nil || !found || again.Leaves[0] != 1 {
		t.Fatal("lookup ownership", again, found, err)
	}
	if err = table.SetBestCut(7, nil); err != nil {
		t.Fatal(err)
	}
	if _, found, err = table.BestCut(7); err != nil || found {
		t.Fatal("clear", found, err)
	}
	mapping := &CNFAreaFlowMapping{Cuts: []AigCNFCut4{{Root: 4, Leaves: []AigID{1, 2}, Truth: 0x8888}}}
	selected, found := mapping.BestCut4(4)
	selected.Leaves[0] = 3
	if !found || mapping.Cuts[0].Leaves[0] != 1 {
		t.Fatal("selected ownership")
	}
	if _, found = mapping.BestCut4(3); found {
		t.Fatal("unexpected cut")
	}
}

func TestCNFHeaderClauseAndLeafIteration(t *testing.T) {
	clauses := [][]SatLit{{0, 3}, {}, {4}}
	stream, err := BuildCNFClauseStream(clauses, CNFHeaderOptions{})
	if err != nil || !slices.Equal(stream.Literals, []SatLit{0, 3, 4}) || !slices.Equal(stream.Ranges, []CNFClauseRange{{0, 2}, {2, 2}, {2, 3}}) {
		t.Fatal(stream, err)
	}
	clauses[0][0] = 8
	if stream.Literals[0] != 0 {
		t.Fatal("stream retained input")
	}
	graph := NewAigManager()
	a, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	b, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	cut := AigCNFCut{Root: 3, Leaves: []AigID{a.ID(), b.ID()}, Truth: headerTruth(t, 2, 0x8888888888888888)}
	objects, err := CNFCutLeafObjects(graph, cut, CNFHeaderOptions{})
	if err != nil || len(objects) != 2 || objects[0].Type != AigCI || objects[1].Type != AigCI {
		t.Fatal(objects, err)
	}
}

func TestCNFHeaderErrorsLimitsAndCancellation(t *testing.T) {
	if _, err := NewCNFBestCutTable(-1, CNFHeaderOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewCNFBestCutTable(2, CNFHeaderOptions{MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	table, err := NewCNFBestCutTable(2, CNFHeaderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	bad := AigCNFCut{Root: 1, Leaves: []AigID{1}, Truth: headerTruth(t, 1, 0xaaaaaaaaaaaaaaaa)}
	if err = table.SetBestCut(1, &bad); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err = table.BestCut(2); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = BuildCNFClauseStream([][]SatLit{{0}, {1}}, CNFHeaderOptions{MaxClauses: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = BuildCNFClauseStream([][]SatLit{{0, 1}}, CNFHeaderOptions{MaxLiterals: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	stop := func() bool { return true }
	if _, err = BuildCNFClauseStream(nil, CNFHeaderOptions{Stop: stop}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err = CNFCutLeafObjects(nil, AigCNFCut{}, CNFHeaderOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	calls := 0
	stopAfterOne := func() bool { calls++; return calls > 2 }
	if _, err = BuildCNFClauseStream([][]SatLit{{0}, {2}}, CNFHeaderOptions{Stop: stopAfterOne}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func runCNFHeaderModel(t testing.TB, seed uint64, steps int) {
	t.Helper()
	table, err := NewCNFBestCutTable(33, CNFHeaderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	present := make([]bool, 33)
	leaves := make([][]AigID, 33)
	for step := 0; step < steps; step++ {
		seed = seed*6364136223846793005 + 1442695040888963407
		root := AigID(1 + seed%32)
		if seed>>8&3 == 0 {
			if err = table.SetBestCut(root, nil); err != nil {
				t.Fatal(err)
			}
			present[root] = false
			leaves[root] = nil
		} else {
			leaf := AigID(seed % uint64(root))
			cut := AigCNFCut{Root: root, Leaves: []AigID{leaf}, Truth: headerTruth(t, 1, 0xaaaaaaaaaaaaaaaa)}
			if err = table.SetBestCut(root, &cut); err != nil {
				t.Fatal(err)
			}
			present[root] = true
			leaves[root] = []AigID{leaf}
		}
		got, found, err := table.BestCut(root)
		if err != nil || found != present[root] || found && !slices.Equal(got.Leaves, leaves[root]) {
			t.Fatal(step, root, got, found, err)
		}
	}
}

func TestCNFHeaderModelCampaign(t *testing.T) {
	for i := 0; i < 5000; i++ {
		runCNFHeaderModel(t, uint64(i+1), 16)
	}
}

func FuzzCNFHeaderBestCuts(f *testing.F) {
	for _, seed := range []uint64{0, 1, 7, 42, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) { runCNFHeaderModel(t, seed, 32) })
}
