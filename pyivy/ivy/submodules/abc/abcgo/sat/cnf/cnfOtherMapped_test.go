package cnf

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

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
		values := make([]bool, len(g.Nodes()))
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
			values := make([]bool, len(g.Nodes()))
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
		vals := make([]bool, len(g.Nodes()))
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
