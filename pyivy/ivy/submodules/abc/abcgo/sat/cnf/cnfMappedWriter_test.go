package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestMappedWriterSourceVariableOrderAndTruth(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	n := g.AND(a, b)
	out := g.AddOutput(n)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	m.selectedCuts = make([]*Cnf_Cut_t, g.NodeSlots())
	m.selectedCuts[n.NodeID()] = Cnf_CutFromTruth([]int32{int32(a.NodeID()), int32(b.NodeID())}, []uint32{0x88888888})
	mapped := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(mapped, n)
	data := Cnf_ManWriteCnf(m, mapped, 1)
	if data == nil {
		t.Fatal("mapped writer failed")
	}
	for _, entry := range []struct {
		edge aig.Edge
		want int32
	}{{out, 1}, {n, 2}, {a, 3}, {b, 4}, {g.Const1(), 5}} {
		if got := data.VarNums[entry.edge.NodeID()]; got != entry.want {
			t.Fatalf("edge %v variable %d want %d", entry.edge, got, entry.want)
		}
	}
	solver := Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil {
		t.Fatal("failed to load CNF")
	}
	for mask := 0; mask < 4; mask++ {
		assum := []bsat.Literal{bsat.Positive(3), bsat.Positive(4)}
		if mask&1 == 0 {
			assum[0] ^= 1
		}
		if mask&2 == 0 {
			assum[1] ^= 1
		}
		want := mask == 3
		lit := bsat.Positive(1)
		if !want {
			lit ^= 1
		}
		if solver.Solve(append(append([]bsat.Literal{}, assum...), lit)...) != bsat.Satisfiable || solver.Solve(append(append([]bsat.Literal{}, assum...), lit^1)...) != bsat.Unsatisfiable {
			t.Fatalf("assignment %d", mask)
		}
	}
}

// The C writer switches from its four-variable SOP table to packed ISOPs
// for a five-leaf cut, preserving the selected output and PI variable order.
func TestMappedWriterFiveInputISOP(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 5)
	leaves := make([]int32, 5)
	for i := range inputs {
		inputs[i] = g.NewInput()
		leaves[i] = int32(inputs[i].NodeID())
	}
	root := inputs[0]
	for _, in := range inputs[1:] {
		root = g.AND(root, in)
	}
	out := g.AddOutput(root)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	m.selectedCuts = make([]*Cnf_Cut_t, g.NodeSlots())
	cut := Cnf_CutFromTruth(leaves, []uint32{0x80000000})
	cut.vIsop[0] = vec.Vec_IntAlloc(5)
	cut.vIsop[1] = vec.Vec_IntAlloc(1)
	var allOnes int32
	for i := 0; i < 5; i++ {
		vec.Vec_IntPush(cut.vIsop[0], 1<<(2*i))
		allOnes |= 2 << (2 * i)
	}
	vec.Vec_IntPush(cut.vIsop[1], allOnes)
	m.selectedCuts[root.NodeID()] = cut
	mapped := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(mapped, root)
	data := Cnf_ManWriteCnf(m, mapped, 1)
	if data == nil {
		t.Fatal("five-input mapped writer failed")
	}
	alternate := Cnf_ManWriteCnfOther(m, mapped)
	if alternate == nil || alternate.Obj2Count[root.NodeID()] != 6 || alternate.VarNums[root.NodeID()] != int32(root.NodeID()) {
		t.Fatal("alternate writer lost the five-input ISOP")
	}
	checkMappedAndOutput(t, alternate, inputs, out)
	checkMappedAndOutput(t, data, inputs, out)
}

func checkMappedAndOutput(t *testing.T, data *Cnf_Dat_t, inputs []aig.Edge, out aig.Edge) {
	t.Helper()
	solver := Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil {
		t.Fatal("failed to load CNF")
	}
	for bits := 0; bits < 32; bits++ {
		assum := make([]bsat.Literal, 0, 6)
		for i, in := range inputs {
			lit := bsat.Positive(data.VarNums[in.NodeID()])
			if bits&(1<<i) == 0 {
				lit ^= 1
			}
			assum = append(assum, lit)
		}
		lit := bsat.Positive(data.VarNums[out.NodeID()])
		if bits != 31 {
			lit ^= 1
		}
		if solver.Solve(append(append([]bsat.Literal(nil), assum...), lit)...) != bsat.Satisfiable || solver.Solve(append(append([]bsat.Literal(nil), assum...), lit^1)...) != bsat.Unsatisfiable {
			t.Fatalf("assignment %d", bits)
		}
	}
}

func TestMappedWriterSourceFreeOutputRestrictions(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	g.AddOutput(a)
	g.AddOutput(a.Not())
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	mapped := vec.Vec_PtrAlloc(0)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("partial combinational outputs accepted")
			}
		}()
		Cnf_ManWriteCnf(m, mapped, 1)
	}()
	if Cnf_ManWriteCnf(m, mapped, 2) == nil {
		t.Fatal("all combinational outputs rejected")
	}
	aig.Aig_ManSetRegNum(g, 1)
	data := Cnf_ManWriteCnf(m, mapped, 1)
	if data == nil || data.VarNums[g.Outputs()[0].NodeID()] != -1 || data.VarNums[g.Outputs()[1].NodeID()] != 1 {
		t.Fatal("last register output was not assigned first")
	}
	if Cnf_ManWriteCnf(m, mapped, 2) == nil {
		t.Fatal("PDR all-output request rejected")
	}
}

func TestMappedWriterSharesStorageWithoutAliasingClauses(t *testing.T) {
	const gates = 128
	m := mappedAndChain(gates)
	defer Cnf_ManStop(m)
	mapped := Cnf_ManScanMapping(m, 1, 1)
	for _, write := range []func() *Cnf_Dat_t{
		func() *Cnf_Dat_t { return Cnf_ManWriteCnf(m, mapped, 0) },
		func() *Cnf_Dat_t { return Cnf_ManWriteCnfOther(m, mapped) },
	} {
		if allocs := testing.AllocsPerRun(3, func() { Cnf_DataFree(write()) }); allocs > 40 {
			t.Errorf("writing %d gates allocated %.0f objects, want at most 40", gates, allocs)
		}
		data := write()
		for i := 0; i < len(data.Clauses)-1; i++ {
			next := data.Clauses[i+1][0]
			private := data.pClauses[i][0]
			data.Clauses[i][0]++
			grown := append(data.Clauses[i], -1)
			if grown[len(grown)-1] != -1 || data.Clauses[i+1][0] != next || data.pClauses[i][0] != private {
				t.Fatalf("clause %d shares mutable storage with another clause or its private copy", i)
			}
		}
	}
}

func TestMappedWriterReadsCompactTruthAcrossAllTwoInputFunctions(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	out := g.AddOutput(root)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	m.selectedCuts = make([]*Cnf_Cut_t, aigObjectCount(g))
	mapped := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(mapped, root)
	leaves := []int32{int32(a.NodeID()), int32(b.NodeID())}
	for truth := uint32(0); truth < 16; truth++ {
		Cnf_ObjSetBestCut(m, &root, Cnf_CutFromTruth(leaves, []uint32{truth}))
		for _, write := range []func() *Cnf_Dat_t{
			func() *Cnf_Dat_t { return Cnf_ManWriteCnf(m, mapped, 1) },
			func() *Cnf_Dat_t { return Cnf_ManWriteCnfOther(m, mapped) },
		} {
			cnf := write()
			solver := Cnf_DataWriteIntoSolver(cnf, 1, 0)
			if solver == nil {
				t.Fatalf("truth %x made free output unsatisfiable", truth)
			}
			for row := range 4 {
				av := bsat.Positive(cnf.VarNums[a.NodeID()])
				bv := bsat.Positive(cnf.VarNums[b.NodeID()])
				y := bsat.Positive(cnf.VarNums[out.NodeID()])
				if row&1 == 0 {
					av ^= 1
				}
				if row&2 == 0 {
					bv ^= 1
				}
				if truth&(1<<row) == 0 {
					y ^= 1
				}
				if solver.Solve(av, bv, y) != bsat.Satisfiable || solver.Solve(av, bv, y^1) != bsat.Unsatisfiable {
					t.Fatalf("compact truth %x, row %d, clauses %v", truth, row, cnf.Clauses)
				}
			}
		}
		mapping := Cnf_ManWriteCnfMapping(m, mapped)
		offset := vec.Vec_IntEntry(mapping, root.NodeID())
		if got := uint32(vec.Vec_IntEntry(mapping, int(offset))); got != truth*0x11111111 {
			t.Errorf("packed mapping truth %08x, want %08x", got, truth*0x11111111)
		}
	}
}

func TestSetBestCutOnFreshManagerAndGrowingGraph(t *testing.T) {
	g := aig.NewGraph()
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	a, b := g.NewInput(), g.NewInput()
	first := g.AND(a, b)
	cut := Cnf_CutFromTruth([]int32{int32(a.NodeID()), int32(b.NodeID())}, []uint32{0x88888888})
	Cnf_ObjSetBestCut(m, &first, cut)
	if got := Cnf_ObjBestCut(m, &first); got != cut {
		t.Error("lost cut installed on fresh manager")
	}
	second := g.AND(a, first)
	next := Cnf_CutFromTruth([]int32{int32(a.NodeID()), int32(first.NodeID())}, []uint32{0x88888888})
	Cnf_ObjSetBestCut(m, &second, next)
	if Cnf_ObjBestCut(m, &first) != cut || Cnf_ObjBestCut(m, &second) != next {
		t.Error("lost cuts after graph grew")
	}
}
