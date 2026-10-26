package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestAigObjectCNFNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/dar_cuts.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []struct {
		Seed, Skip int
		Ops        [][2]uint32
		Outputs    []uint32
		Other      *struct {
			Variables      int
			Starts, Counts []int
			Clauses        [][]SatLit
			Mapping        []int32
		}
	}
	if e = json.Unmarshal(data, &cases); e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, row := range cases {
		if row.Other == nil {
			continue
		}
		count++
		g := darBuildGraph(t, row.Ops, row.Outputs)
		c, e := DeriveAigObjectCNF4(g, AigCNFDeriveOptions{Cuts: DARCutOptions{SkipTruthMinimization: row.Skip != 0}})
		if e != nil {
			t.Fatal(row.Seed, row.Skip, e)
		}
		if c.variables != row.Other.Variables || !slices.Equal(c.starts, row.Other.Starts) || !slices.Equal(c.counts, row.Other.Counts) || !slices.Equal(c.mapping, row.Other.Mapping) || len(c.clauses) != len(row.Other.Clauses) {
			t.Fatal(row.Seed, row.Skip, c, row.Other)
		}
		for i, clause := range c.clauses {
			if !slices.Equal(clause, row.Other.Clauses[i]) {
				t.Fatal(row.Seed, row.Skip, i, clause, row.Other.Clauses[i])
			}
		}
		checkDerivedCNFSemantics(t, g, c.AigMappedCNF)
	}
	if count != 48 {
		t.Fatal(count)
	}
}
func TestAigObjectCNFLimitsAndSequential(t *testing.T) {
	g := mappedCNFFixture(t, 2, 1)
	cuts := []AigCNFCut4{{Root: 7, Leaves: []AigID{1, 2, 3, 4}, Truth: 0x8000}}
	c, e := WriteAigObjectCNF4(g, cuts, MappedCNFOptions{MaxVariables: len(g.objects)})
	if e != nil {
		t.Fatal(e)
	}
	if c.variables != len(g.objects) || c.clauses[len(c.clauses)-1][0] != 0 || c.counts[0] != 1 || c.counts[1] != 0 || c.starts[1] != -1 || c.counts[5] != -1 || c.counts[g.cos[0]] != 2 {
		t.Fatal(c)
	}
	checkMappedCNFProjection(t, g, cuts, c.AigMappedCNF)
	starts := c.ObjectClauseStarts()
	starts[0] = 99
	if c.starts[0] == 99 {
		t.Fatal("alias")
	}
	counts := c.ObjectClauseCounts()
	counts[0] = 99
	if c.counts[0] == 99 {
		t.Fatal("alias")
	}
	for _, o := range []MappedCNFOptions{{MaxVariables: len(g.objects) - 1}, {MaxClauses: 1}, {MaxLiterals: 1}, {Stop: func() bool { return true }}} {
		r, e := WriteAigObjectCNF4(g, cuts, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	empty := NewAigManager()
	r, e := WriteAigObjectCNF4(empty, nil, MappedCNFOptions{MaxVariables: 1})
	if e != nil || r.variables != 1 || len(r.clauses) != 1 || r.clauses[0][0] != 0 || r.starts[0] != 0 || r.counts[0] != 1 {
		t.Fatal(r, e)
	}
}
func FuzzAigObjectCNF(f *testing.F) {
	f.Add([]byte{2, 4, 6, 8, 14, 16}, false)
	f.Fuzz(func(t *testing.T, data []byte, skip bool) {
		if len(data) > 64 {
			data = data[:64]
		}
		var ops [][2]uint32
		for i := 0; i+1 < len(data); i += 2 {
			ops = append(ops, [2]uint32{uint32(data[i]), uint32(data[i+1])})
		}
		g := darBuildGraph(t, ops, []uint32{uint32(2 * (6 + len(ops))), 3})
		c, e := DeriveAigObjectCNF4(g, AigCNFDeriveOptions{Cuts: DARCutOptions{SkipTruthMinimization: skip}})
		if e != nil {
			t.Fatal(e)
		}
		checkDerivedCNFSemantics(t, g, c.AigMappedCNF)
		for id, start := range c.starts {
			if start < 0 {
				continue
			}
			if start+c.counts[id] > len(c.clauses) {
				t.Fatal(id, c)
			}
			for _, clause := range c.clauses[start : start+c.counts[id]] {
				if int(clause[0].Var()) != id {
					t.Fatal(id, clause)
				}
			}
		}
	})
}

func TestAigObjectCNFWideProjection(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_mapped_wide.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Inputs, OutputMode, Phase int
		Truth                     []uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	for _, row := range rows {
		g, root := wideCNFGraph(t, row.Inputs, row.OutputMode, row.Phase)
		truth, e := TruthTableFromWords(row.Inputs, row.Truth, 0)
		if e != nil {
			t.Fatal(e)
		}
		c, e := WriteAigObjectCNF(g, []AigCNFCut{{Root: root, Leaves: g.Inputs(), Truth: truth}}, MappedCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if c.starts[root] != 0 || c.counts[root] != len(c.clauses)-5 || c.starts[0] != len(c.clauses)-1 {
			t.Fatal(c)
		}
		checkWideProjection(t, g, c.AigMappedCNF, truth, row.Phase)
	}
}
