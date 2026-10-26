package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func equalSatClauses(a, b [][]SatLit) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !slices.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
func TestGiaWindowCNFNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_window_cnf.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		NodeCount, First, Variables int
		Ranges                      []GiaWindowCutRange
		Cuts                        []GiaWindowCut
		Clauses                     [][]SatLit
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 24 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		r, err := BuildGiaWindowCNF(row.NodeCount, row.First, row.Cuts, row.Ranges, GiaWindowCNFOptions{})
		if err != nil || r.Variables != row.Variables || r.FirstVariable != row.First || !equalSatClauses(r.Clauses, row.Clauses) {
			t.Fatal(index, r, err)
		}
	}
}
func satClausesHold(clauses [][]SatLit, assignment uint64) bool {
	for _, clause := range clauses {
		holds := false
		for _, lit := range clause {
			value := assignment>>uint(lit.Var())&1 != 0
			if lit.Negative() {
				value = !value
			}
			holds = holds || value
		}
		if !holds {
			return false
		}
	}
	return true
}
func TestGiaWindowCNFMeaningAndSolver(t *testing.T) {
	cuts := []GiaWindowCut{{Node: 0}, {Node: -1, Nodes: [2]uint64{0, 1 << 63}}, {Node: 1, Nodes: [2]uint64{1, 0}}, {Node: -1}}
	ranges := []GiaWindowCutRange{{Start: 0, Count: 2}, {Start: 2, Count: 2}}
	r, err := BuildGiaWindowCNF(2, 2, cuts, ranges, GiaWindowCNFOptions{MaxVariables: 6, MaxClauses: 5, MaxLiterals: 10})
	if err != nil || !equalSatClauses(r.Clauses, [][]SatLit{{1, 4}, {5, 0}, {3, 8}, {9, 2}, {9, 0}}) {
		t.Fatal(r, err)
	}
	for assignment := uint64(0); assignment < 64; assignment++ {
		node0, node1 := assignment&1 != 0, assignment&2 != 0
		cut0, cut2 := assignment&4 != 0, assignment&16 != 0
		expected := (!node0 || cut0) && (!cut0 || node0) && (!node1 || cut2) && (!cut2 || node1 && node0)
		if satClausesHold(r.Clauses, assignment) != expected {
			t.Fatal("declarative closure", assignment)
		}
	}
	multiple, err := BuildGiaWindowCNF(1, 1, make([]GiaWindowCut, 3), []GiaWindowCutRange{{Count: 3}}, GiaWindowCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	solver := NewSatSolver()
	solver.SetVariableCount(multiple.Variables)
	for _, clause := range multiple.Clauses {
		if _, err := solver.AddClause(clause); err != nil {
			t.Fatal(err)
		}
	}
	answer, err := solver.Solve([]SatLit{0, 2, 4}, SatLimits{})
	if err != nil || answer.Status != SatSatisfiable {
		t.Fatal("native permits multiple cuts", answer, err)
	}
	unit, err := BuildGiaWindowCNF(1, 1, []GiaWindowCut{{Nodes: [2]uint64{0, 1 << 63}}}, []GiaWindowCutRange{{Count: 1}}, GiaWindowCNFOptions{})
	if err != nil || !equalSatClauses(unit.Clauses, [][]SatLit{{1}}) {
		t.Fatal("final unit excluded", unit, err)
	}
}
func TestGiaWindowCNFLimits(t *testing.T) {
	cuts := []GiaWindowCut{{}, {}, {Nodes: [2]uint64{1, 0}}, {}}
	ranges := []GiaWindowCutRange{{Count: 2}, {Start: 2, Count: 2}}
	for _, options := range []GiaWindowCNFOptions{{MaxVariables: 5}, {MaxClauses: 4}, {MaxLiterals: 9}, {MaxVariables: -1}, {MaxClauses: -1}, {MaxLiterals: -1}, {Stop: func() bool { return true }}} {
		if r, err := BuildGiaWindowCNF(2, 2, cuts, ranges, options); r.Clauses != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	calls := 0
	if r, err := BuildGiaWindowCNF(2, 2, cuts, ranges, GiaWindowCNFOptions{Stop: func() bool { calls++; return calls >= 7 }}); r.Clauses != nil || err == nil {
		t.Fatal(r, err)
	}
	for _, bad := range []GiaWindowCutRange{{Start: -1, Count: 1}, {Count: 0}, {Start: 4, Count: 1}, {Count: 5}} {
		if _, err := BuildGiaWindowCNF(1, 1, cuts, []GiaWindowCutRange{bad}, GiaWindowCNFOptions{}); err == nil {
			t.Fatal(bad)
		}
	}
	invalid := []GiaWindowCut{{Nodes: [2]uint64{0, 1 << 63}}, {}}
	if _, err := BuildGiaWindowCNF(1, 1, invalid, []GiaWindowCutRange{{Count: 2}}, GiaWindowCNFOptions{}); err == nil {
		t.Fatal("invalid internal-node mask")
	}
}
func FuzzGiaWindowCNF(f *testing.F) {
	f.Add(uint8(0), uint64(1))
	f.Add(uint8(255), uint64(0xabcdef123456789))
	f.Fuzz(func(t *testing.T, flags uint8, seed uint64) {
		n := 1 + int(flags%3)
		first := n + int(flags>>4)%3
		cuts := make([]GiaWindowCut, 0)
		ranges := make([]GiaWindowCutRange, n)
		next := func() uint64 { seed ^= seed << 13; seed ^= seed >> 7; seed ^= seed << 17; return seed }
		for i := range ranges {
			proper := int(next() % 3)
			ranges[i] = GiaWindowCutRange{Start: len(cuts), Count: proper + 1}
			for j := 0; j <= proper; j++ {
				cuts = append(cuts, GiaWindowCut{Node: i, Nodes: [2]uint64{next() & ((uint64(1) << uint(n)) - 1), 0}})
			}
		}
		r, err := BuildGiaWindowCNF(n, first, cuts, ranges, GiaWindowCNFOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for sample := 0; sample < 34; sample++ {
			assignment := next()
			if sample == 0 {
				assignment = 0
			}
			if sample == 1 {
				assignment = ^uint64(0)
			}
			expected := true
			for node, span := range ranges {
				nodeSelected := assignment>>uint(node)&1 != 0
				anyCut := false
				for index := span.Start; index < span.Start+span.Count-1; index++ {
					selected := assignment>>uint(first+index)&1 != 0
					anyCut = anyCut || selected
					if selected {
						if !nodeSelected {
							expected = false
						}
						for fanin := 0; fanin < n; fanin++ {
							if cuts[index].Nodes[0]>>uint(fanin)&1 != 0 && assignment>>uint(fanin)&1 == 0 {
								expected = false
							}
						}
					}
				}
				if nodeSelected && !anyCut {
					expected = false
				}
			}
			if satClausesHold(r.Clauses, assignment) != expected {
				t.Fatal("selection/closure predicate", assignment, r.Clauses)
			}
		}
	})
}
