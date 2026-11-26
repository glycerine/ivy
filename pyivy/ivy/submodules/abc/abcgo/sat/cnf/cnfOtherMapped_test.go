package cnf

import (
	"reflect"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestMappedOutputDrivenByInvertingBuffer(t *testing.T) {
	g := aig.NewGraph()
	input := g.NewInput()
	buf := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{input.Not(), aig.InvalidEdge}})
	output := g.AddOutput(buf)
	for _, d := range []*Cnf_Dat_t{Cnf_DeriveOther(g, 0), Cnf_Derive(g, 1)} {
		if d == nil {
			t.Error("missing CNF for buffered output")
			continue
		}
		s := Cnf_DataWriteIntoSolver(d, 1, 0)
		x, y := bsat.Positive(d.VarNums[input.NodeID()]), bsat.Positive(d.VarNums[output.NodeID()])
		for _, value := range []bsat.Literal{0, 1} {
			if s.Solve(x^value, y^value^1) != bsat.Satisfiable || s.Solve(x^value, y^value) != bsat.Unsatisfiable {
				t.Fatalf("buffer lost polarity at literal sign %d", value)
			}
		}
	}
}

// The original C Cnf_DeriveOther produced these arrays for !(a & !b & c)
// with both values of fSkipTtMin. Gate 4 is absorbed into the cut of gate 5.
func TestAlternateEntryPointsMatchNativeCuts(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(g.AND(a, b.Not()), c)
	g.AddOutput(root.Not())
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	wantClauses := [][]int32{{10, 3, 4, 7}, {11, 2}, {11, 5}, {11, 6}, {12, 10}, {13, 11}, {0}}
	wantMapping := []int32{0, 0, 0, 0, 0, 7, 0, 538976288, 1, 2, 3, -1}
	for _, skip := range []int32{0, 1} {
		for _, got := range []*Cnf_Dat_t{Cnf_DeriveOther(g, skip), Cnf_DeriveOtherWithMan(m, g, skip)} {
			if got == nil {
				t.Fatal("missing alternate CNF")
			}
			if got.Variables != 7 || !reflect.DeepEqual(got.Clauses, wantClauses) || !slices.Equal(got.Mapping, wantMapping) {
				t.Errorf("skip %d: %d variables, clauses %v, mapping %v", skip, got.Variables, got.Clauses, got.Mapping)
			}
			if !slices.Equal(got.Obj2Clause, []int32{6, -1, -1, -1, -1, 0, 4}) || !slices.Equal(got.Obj2Count, []int32{1, 0, 0, 0, -1, 4, 2}) {
				t.Errorf("skip %d: object clause starts %v, counts %v", skip, got.Obj2Clause, got.Obj2Count)
			}
		}
	}
	m.pManAig = g
	Cnf_DeriveMapping(m)
	mapped := Cnf_ManScanMapping(m, 1, 1)
	data := Cnf_ManWriteCnfOther(m, mapped)
	if data == nil || !reflect.DeepEqual(data.Clauses, wantClauses) {
		t.Fatal("alternate writer omitted the supplied cuts or constrained the output")
	}
}

func TestOtherMappedCNFExhaustive(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	left := g.AND(a, b)
	root := g.AND(left, c.Not())
	out := g.AddOutput(root)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	data, err := deriveOtherMapped(m, g, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Mapping) == 0 || data.Obj2Count[0] != 1 || data.Obj2Count[out.NodeID()] != 2 {
		t.Fatalf("missing mapping or clauses: %+v", data)
	}
	offset := data.Mapping[root.NodeID()]
	if offset == 0 {
		t.Fatal("root cut not mapped")
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := make([]bool, g.NodeSlots())
		values[0] = true
		values[a.NodeID()] = assignment&1 != 0
		values[b.NodeID()] = assignment&2 != 0
		values[c.NodeID()] = assignment&4 != 0
		values[left.NodeID()] = values[a.NodeID()] && values[b.NodeID()]
		values[root.NodeID()] = values[left.NodeID()] && !values[c.NodeID()]
		values[out.NodeID()] = values[root.NodeID()]
		satisfied := func(clause []int32) bool {
			for _, lit := range clause {
				v := values[lit>>1]
				if lit&1 != 0 {
					v = !v
				}
				if v {
					return true
				}
			}
			return false
		}
		for _, clause := range data.Clauses {
			if !satisfied(clause) {
				t.Fatalf("assignment=%d invalid clause %v", assignment, clause)
			}
		}
		for id, count := range data.Obj2Count {
			if count <= 0 {
				continue
			}
			values[id] = !values[id]
			broken := false
			for _, clause := range data.Clauses[data.Obj2Clause[id] : data.Obj2Clause[id]+count] {
				if !satisfied(clause) {
					broken = true
					break
				}
			}
			values[id] = !values[id]
			if !broken {
				t.Fatalf("assignment=%d object=%d not constrained", assignment, id)
			}
		}
	}
	if len(vec.Vec_IntArray(data.vMapping)) != len(data.Mapping) {
		t.Fatal("mapping views differ")
	}
}

func TestOtherMappedCNFComplementedCuts(t *testing.T) {
	for mask := 0; mask < 8; mask++ {
		g := aig.NewGraph()
		inputs := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		edge := func(i int) aig.Edge {
			if mask&(1<<i) != 0 {
				return inputs[i].Not()
			}
			return inputs[i]
		}
		root := g.AND(g.AND(edge(0), edge(1)), edge(2))
		out := g.AddOutput(root)
		data, err := Cnf_DeriveOtherMapped(g, 1)
		if err != nil {
			t.Fatalf("mask=%d: %v", mask, err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			values := make([]bool, g.NodeSlots())
			values[0] = true
			for i, input := range inputs {
				values[input.NodeID()] = assignment&(1<<i) != 0
			}
			for id, node := range g.Nodes() {
				get := func(e aig.Edge) bool {
					v := values[e.NodeID()]
					if e.Complemented() {
						return !v
					}
					return v
				}
				if node.Kind == aig.And {
					values[id] = get(node.Fanins[0]) && get(node.Fanins[1])
				} else if node.Kind == aig.Output {
					values[id] = get(node.Fanins[0])
				}
			}
			for _, clause := range data.Clauses {
				sat := false
				for _, lit := range clause {
					v := values[lit>>1]
					if lit&1 != 0 {
						v = !v
					}
					sat = sat || v
				}
				if !sat {
					t.Fatalf("mask=%d assignment=%d clause=%v output=%v", mask, assignment, clause, values[out.NodeID()])
				}
			}
		}
	}
}

func TestOtherMappedCNFXor(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x := g.NewXor(a, b)
	out := g.AddOutput(x)
	data, err := Cnf_DeriveOtherMapped(g, 0)
	if err != nil {
		t.Fatal(err)
	}
	if data.Obj2Count[x.NodeID()] != 4 {
		t.Fatalf("XOR clause count=%d", data.Obj2Count[x.NodeID()])
	}
	for assignment := 0; assignment < 4; assignment++ {
		vals := make([]bool, g.NodeSlots())
		vals[0] = true
		vals[a.NodeID()] = assignment&1 != 0
		vals[b.NodeID()] = assignment&2 != 0
		vals[x.NodeID()] = vals[a.NodeID()] != vals[b.NodeID()]
		vals[out.NodeID()] = vals[x.NodeID()]
		for _, clause := range data.Clauses {
			sat := false
			for _, lit := range clause {
				v := vals[lit>>1]
				if lit&1 != 0 {
					v = !v
				}
				sat = sat || v
			}
			if !sat {
				t.Fatalf("assignment=%d clause=%v", assignment, clause)
			}
		}
	}
}
