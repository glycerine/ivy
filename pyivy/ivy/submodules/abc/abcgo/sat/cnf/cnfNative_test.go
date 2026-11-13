package cnf

import (
	"math/rand"
	"os"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestEncodeAIGRandomGraphsAgainstEvaluation(t *testing.T) {
	rng := rand.New(rand.NewSource(70814))
	for trial := 0; trial < 100; trial++ {
		g := aig.NewGraph()
		pool := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.Const0(), g.Const1()}
		for i := 0; i < 7; i++ {
			a, b := pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))]
			if rng.Intn(2) == 0 {
				a = a.Not()
			}
			if rng.Intn(2) == 0 {
				b = b.Not()
			}
			if rng.Intn(2) == 0 {
				pool = append(pool, g.AND(a, b))
			} else {
				pool = append(pool, g.NewXor(a, b))
			}
		}
		for i := 0; i < 2; i++ {
			root := pool[rng.Intn(len(pool))]
			if rng.Intn(2) == 0 {
				root = root.Not()
			}
			g.AddOutput(root)
		}
		for outputCount := 0; outputCount <= 2; outputCount++ {
			data, err := EncodeAIG(g, outputCount)
			if err != nil {
				t.Fatal(err)
			}
			s := Cnf_DataWriteIntoSolver(data, 1, 0)
			for mask := 0; mask < 8; mask++ {
				values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				outputs, err := g.Evaluate(values)
				if err != nil {
					t.Fatal(err)
				}
				want := true
				for _, value := range outputs[:outputCount] {
					want = want && value
				}
				got := false
				if s != nil {
					var assumptions []bsat.Literal
					for i, edge := range g.Inputs() {
						variable := data.VarNums[edge.NodeID()]
						if values[i] {
							assumptions = append(assumptions, bsat.Positive(variable))
						} else {
							assumptions = append(assumptions, bsat.Negative(variable))
						}
					}
					got = s.Solve(assumptions...) == bsat.Satisfiable
				}
				if got != want {
					t.Fatalf("trial=%d outputs=%d inputs=%03b got=%v want=%v", trial, outputCount, mask, got, want)
				}
			}
		}
	}
}

func TestEncodeExplicitXorOfConstants(t *testing.T) {
	for _, a := range []bool{false, true} {
		for _, b := range []bool{false, true} {
			g := aig.NewGraph()
			left, right := g.Const0(), g.Const0()
			if a {
				left = g.Const1()
			}
			if b {
				right = g.Const1()
			}
			root := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.ExclusiveOr, Fanins: [2]aig.Edge{left, right}})
			g.AddOutput(root)
			data, err := EncodeAIG(g, 1)
			if err != nil {
				t.Fatal(err)
			}
			got := Cnf_DataWriteIntoSolver(data, 1, 0) != nil
			if got != (a != b) {
				t.Fatalf("%v xor %v: satisfiable=%v", a, b, got)
			}
		}
	}
}

func TestAlternateDerivationRecordsObjectClauses(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x := g.NewXor(a, b)
	out := g.AddOutput(x)
	data := Cnf_DeriveOther(g, 0)
	if data == nil || len(data.Obj2Clause) != len(g.Nodes()) || len(data.Obj2Count) != len(g.Nodes()) {
		t.Fatalf("missing object clause metadata: %+v", data)
	}
	for _, edge := range []aig.Edge{x, out} {
		id := edge.NodeID()
		start, count := data.Obj2Clause[id], data.Obj2Count[id]
		if start < 0 || count == 0 || int(start+count) > len(data.Clauses) {
			t.Fatalf("object %d range %d+%d", id, start, count)
		}
		for _, clause := range data.Clauses[start : start+count] {
			if len(clause) == 0 || clause[0]>>1 != data.VarNums[id] {
				t.Fatalf("object %d clause %v lacks output literal first", id, clause)
			}
		}
	}
}

func TestDeriveMappingArrayUsesObjectOffsets(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	root := g.AND(x, y)
	g.AddOutput(root)
	mapping := vec.Vec_IntArray(Cnf_DeriveMappingArray(g))
	if len(mapping) < len(g.Nodes()) {
		t.Fatalf("short mapping %v", mapping)
	}
	offset := int(mapping[root.NodeID()])
	if offset < len(g.Nodes()) || offset+5 > len(mapping) {
		t.Fatalf("invalid root offset %d in %v", offset, mapping)
	}
	if uint32(mapping[offset])&15 != 8 {
		t.Fatalf("wrong root truth %#x", mapping[offset])
	}
	if !reflect.DeepEqual(mapping[offset+1:offset+5], []int32{int32(x.NodeID()), int32(y.NodeID()), -1, -1}) {
		t.Fatalf("wrong cut leaves %v", mapping[offset+1:offset+5])
	}
}

func TestCreateSelectedCutCopiesMapping(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	g.AddOutput(root)
	manager := Cnf_ManStart()
	manager.pManAig = g
	Cnf_DeriveMapping(manager)
	selected := Cnf_ObjBestCut(manager, &root)
	created := Cnf_CutCreate(manager, &root)
	if selected == nil || created == nil || created == selected || !reflect.DeepEqual(created.pFanins, selected.pFanins) || !reflect.DeepEqual(created.truth, selected.truth) {
		t.Fatalf("selected=%v created=%v", selected, created)
	}
	created.pFanins[0] = -1
	created.truth[0] = 0
	if selected.pFanins[0] < 0 || selected.truth[0] == 0 {
		t.Fatal("created cut aliases selected mapping")
	}
}

func cnfSatisfiableWithInputs(cnf *CNF, inputs map[int]bool) bool {
	for assignment := 0; assignment < 1<<cnf.Variables; assignment++ {
		matches := true
		for variable, value := range inputs {
			if ((assignment>>variable)&1 != 0) != value {
				matches = false
				break
			}
		}
		if !matches {
			continue
		}
		valid := true
		for _, clause := range cnf.Clauses {
			satisfied := false
			for _, lit := range clause {
				value := (assignment>>(lit>>1))&1 != 0
				if lit&1 != 0 {
					value = !value
				}
				if value {
					satisfied = true
					break
				}
			}
			if !satisfied {
				valid = false
				break
			}
		}
		if valid {
			return true
		}
	}
	return false
}

func TestEncodeAIGAndAssertsOutput(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	and := g.AND(a, b)
	g.AddOutput(and)
	cnf, err := EncodeAIG(g, -1)
	if err != nil {
		t.Fatal(err)
	}
	if err := cnf.Validate(); err != nil {
		t.Fatal(err)
	}
	if cnf.Variables != 4 {
		t.Fatalf("variables=%d, want 4", cnf.Variables)
	}
	if cnf.VarNums[0] != -1 || cnf.VarNums[a.NodeID()] != 0 || cnf.VarNums[b.NodeID()] != 1 {
		t.Fatalf("unexpected object map: %v", cnf.VarNums)
	}
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			want := x == 1 && y == 1
			if got := cnfSatisfiableWithInputs(cnf, map[int]bool{0: x == 1, 1: y == 1}); got != want {
				t.Errorf("inputs %d,%d satisfiable=%t, want %t", x, y, got, want)
			}
		}
	}
}

func TestEncodeAIGXORAndComplement(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	xor := g.XOR(a.Not(), b)
	g.AddOutput(xor)
	cnf, err := EncodeAIG(g, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := cnf.Validate(); err != nil {
		t.Fatal(err)
	}
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			want := (x == 0) != (y == 1)
			if got := cnfSatisfiableWithInputs(cnf, map[int]bool{0: x == 1, 1: y == 1}); got != want {
				t.Errorf("inputs %d,%d satisfiable=%t, want %t", x, y, got, want)
			}
		}
	}
}

func TestEncodeAIGConstantFalseOutput(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const0())
	cnf, err := EncodeAIG(g, 1)
	if err != nil {
		t.Fatal(err)
	}
	if cnfSatisfiableWithInputs(cnf, nil) {
		t.Fatal("constant-false output should make the formula unsatisfiable")
	}
}

func TestCutComposeUsesOwnedTruthTables(t *testing.T) {
	cut := Cnf_CutFromTruth([]int32{1, 2}, []uint32{0x8}) // x AND y
	fan := Cnf_CutFromTruth([]int32{3}, []uint32{0x2})    // z
	composed := Cnf_CutCompose(nil, cut, fan, 2)
	if composed == nil {
		t.Fatal("composition returned nil")
	}
	if got := composed.pFanins; len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("support=%v, want [1 3]", got)
	}
	if got := composed.truth[0]; got != 0x8 {
		t.Fatalf("truth=%#x, want x AND z truth %#x", got, uint32(0x8))
	}
	Cnf_CutFree(cut)
	Cnf_CutFree(fan)
	if composed.truth[0] != 0x8 {
		t.Fatal("result truth aliases freed inputs")
	}
}

func TestDIMACSReadAndSolverBridge(t *testing.T) {
	name := t.TempDir() + "/small.cnf"
	if err := os.WriteFile(name, []byte("c two variables\np cnf 2 2\n1 2 0\n-1 2 0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	data := Cnf_DataReadFromFile(name)
	if data == nil {
		t.Fatal("failed to read DIMACS")
	}
	if data.Variables != 2 || len(data.Clauses) != 2 {
		t.Fatalf("got %d vars, %d clauses", data.Variables, len(data.Clauses))
	}
	solver := Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil {
		t.Fatal("expected satisfiable solver bridge")
	}
	if solver.Solve() != bsat.Satisfiable {
		t.Fatal("bridge changed satisfiable CNF")
	}
}

func TestNativeCutMappingOwnsCopies(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(g.AND(a, b), c)
	g.AddOutput(root)
	manager := Cnf_ManStart()
	manager.pManAig = g
	Cnf_DeriveMapping(manager)
	cut := manager.selectedCuts[root.NodeID()]
	if cut == nil {
		t.Fatal("missing selected cut")
	}
	if len(cut.pFanins) == 0 {
		t.Fatal("selected cut lost its support")
	}
	for _, leaf := range cut.pFanins {
		if leaf == int32(root.NodeID()) {
			t.Fatal("selected cut depends on itself")
		}
	}
	mapping := Cnf_ManScanMapping(manager, 1, 0)
	if vec.Vec_PtrSize(mapping) == 0 {
		t.Fatal("mapping traversal returned no gates")
	}
	Cnf_ManStop(manager)
}

func TestCnfDataAddXorClauseTruthTable(t *testing.T) {
	for a := int32(0); a <= 1; a++ {
		for b := int32(0); b <= 1; b++ {
			for c := int32(0); c <= 1; c++ {
				solver := bsat.NewSolver()
				if Cnf_DataAddXorClause(solver, 0, 1, 2) != 4 {
					t.Fatal("failed to add XOR clauses")
				}
				for variable, value := range []int32{a, b, c} {
					lit := bsat.Negative(int32(variable))
					if value != 0 {
						lit = bsat.Positive(int32(variable))
					}
					solver.AddClause(lit)
				}
				got := solver.Solve() == bsat.Satisfiable
				want := c == (a ^ b)
				if got != want {
					t.Fatalf("a=%d b=%d c=%d: satisfiable=%v, want %v", a, b, c, got, want)
				}
			}
		}
	}
}
