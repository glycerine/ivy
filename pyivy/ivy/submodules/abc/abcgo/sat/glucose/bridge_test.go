package glucose

import (
	"github.com/glycerine/abcgo/aig/aig"
	"strings"
	"testing"
)

func truthGraph(table uint8) (*aig.Graph, aig.Edge) {
	g := aig.NewGraph()
	inputs := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
	root := g.Const0()
	for bits := 0; bits < 8; bits++ {
		if table&(1<<bits) == 0 {
			continue
		}
		cube := g.Const1()
		for i, x := range inputs {
			if bits&(1<<i) == 0 {
				x = x.Not()
			}
			cube = g.AND(cube, x)
		}
		root = g.OR(root, cube)
	}
	g.AddOutput(root)
	return g, root
}
func TestBridgeAIGAllThreeVariableFunctions(t *testing.T) {
	for table := 0; table < 256; table++ {
		g, _ := truthGraph(uint8(table))
		status, model := Glucose_SolveAig(g, &Glucose_Pars{Pre: 1})
		if (status == 10) != (table != 0) {
			t.Fatalf("table%d status%d", table, status)
		}
		if status == 10 {
			bits := 0
			for i, v := range model {
				bits |= int(v) << i
			}
			if table&(1<<bits) == 0 {
				t.Fatal("invalid graph model", table, model)
			}
		}
	}
}
func TestBridgeOverlapMeaning(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	s := NewSolver()
	if Bmcg_sat_solver_equiv_overlap_check(s, g, int32(x), int32(x), 0) != 1 {
		t.Fatal("same input must overlap")
	}
	if Bmcg_sat_solver_equiv_overlap_check(s, g, int32(x), int32(x.Not()), 0) != 0 {
		t.Fatal("opposite inputs must not overlap")
	}
}

func evalGraphEdge(g *aig.Graph, root aig.Edge, bits int) bool {
	input := make(map[int]bool)
	for i, x := range g.Inputs() {
		input[x.NodeID()] = bits&(1<<i) != 0
	}
	memo := make(map[aig.Edge]bool)
	var rec func(aig.Edge) bool
	rec = func(e aig.Edge) bool {
		if value, ok := memo[e]; ok {
			return value
		}
		n := g.Node(e)
		value := false
		switch n.Kind {
		case aig.Constant:
			value = true
		case aig.Input:
			value = input[e.NodeID()]
		case aig.And:
			value = rec(n.Fanins[0]) && rec(n.Fanins[1])
		case aig.ExclusiveOr:
			value = rec(n.Fanins[0]) != rec(n.Fanins[1])
		case aig.Buffer, aig.Output:
			value = rec(n.Fanins[0])
		}
		if e.Complemented() {
			value = !value
		}
		memo[e] = value
		return value
	}
	return rec(root)
}
func evalSOP(sop string, bits, n int) bool {
	for _, row := range strings.Split(strings.TrimSuffix(sop, "\x00"), "\n") {
		if row == "" {
			continue
		}
		if strings.TrimSpace(row) == "1" {
			return true
		}
		if strings.TrimSpace(row) == "0" {
			continue
		}
		if len(row) < n+2 {
			panic("malformed SOP")
		}
		matches := true
		for i := 0; i < n; i++ {
			if row[i] == '1' && bits&(1<<i) == 0 || row[i] == '0' && bits&(1<<i) != 0 {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}
func TestBridgeSOPAndQuantificationExhaustive(t *testing.T) {
	for table := 0; table < 256; table++ {
		g, root := truthGraph(uint8(table))
		sop := Bmcg_sat_solver_sop(g, 0)
		for bits := 0; bits < 8; bits++ {
			if got := evalSOP(sop, bits, 3); got != (table&(1<<bits) != 0) {
				t.Fatalf("SOP table%d bits%d cover%q", table, bits, sop)
			}
		}
		for mask := 0; mask < 8; mask++ {
			keep := func(_ any, i int32) int32 { return boolInt(mask&(1<<i) != 0) }
			var divisors []int32
			projected := Bmcg_sat_solver_quantify([2]*Solver{NewSolver(), NewSolver()}, g, int32(root), 1, keep, nil, &divisors)
			cofactored := Bmcg_sat_solver_quantify2(g, int32(root), 1, keep, nil, nil)
			for bits := 0; bits < 8; bits++ {
				want := false
				for witness := 0; witness < 8; witness++ {
					if witness&mask == bits&mask && table&(1<<witness) != 0 {
						want = true
						break
					}
				}
				if got := evalGraphEdge(g, aig.Edge(projected), bits); got != want {
					t.Fatalf("SAT projection table%d mask%d bits%d got%v want%v", table, mask, bits, got, want)
				}
				if got := evalGraphEdge(g, aig.Edge(cofactored), bits); got != want {
					t.Fatalf("cofactor projection table%d mask%d bits%d", table, mask, bits)
				}
			}
		}
	}
}
func TestBridgeWrapperMinimizationAndControls(t *testing.T) {
	s := Bmcg_sat_solver_start()
	Bmcg_sat_solver_set_nvars(s, 3)
	if Bmcg_sat_solver_addclause(s, []int32{1, 3}, 2) != 1 {
		t.Fatal("add")
	}
	as := []int32{0, 2, 4}
	if Bmcg_sat_solver_solve(s, as, 3) != GLUCOSE_UNSAT {
		t.Fatal("expected assumption conflict")
	}
	n := Bmcg_sat_solver_minimize_assumptions(s, as, 3, 0)
	if n != 2 || Bmcg_sat_solver_solve(s, as, n) != GLUCOSE_UNSAT {
		t.Fatal("minimal core", as, n)
	}
	for i := int32(0); i < n; i++ {
		test := append([]int32(nil), as[:i]...)
		test = append(test, as[i+1:n]...)
		if Bmcg_sat_solver_solve(s, test, int32(len(test))) != GLUCOSE_SAT {
			t.Fatal("core not minimal")
		}
	}
	calls := 0
	Bmcg_sat_solver_setcallback(s, "manager", func(m any, status int32, model []int32) int32 {
		if m != "manager" || status != 1 || len(model) != 3 {
			t.Fatal(m, status, model)
		}
		calls++
		return 0
	})
	if Bmcg_sat_solver_solve(s, nil, 0) != GLUCOSE_SAT || calls != 1 {
		t.Fatal("callback")
	}
	Bmcg_sat_solver_setcallback(s, nil, nil)
	var stop int32 = 1
	Bmcg_sat_solver_set_stop(s, &stop)
	if Bmcg_sat_solver_solve(s, nil, 0) != GLUCOSE_UNDEC {
		t.Fatal("stop flag")
	}
	Bmcg_sat_solver_set_stop(s, nil)
	Bmcg_sat_solver_set_jftr(s, 2)
	Bmcg_sat_solver_set_var_fanin_lit(s, 2, 0, 2)
	Bmcg_sat_solver_start_new_round(s)
	Bmcg_sat_solver_mark_cone(s, 2)
	if Bmcg_sat_solver_jftr(s) != 2 || Bmcg_sat_solver_solve(s, nil, 0) != GLUCOSE_SAT {
		t.Fatal("legacy inert circuit hooks changed CNF")
	}
}
func TestBridgeGateEncodings(t *testing.T) {
	for _, xor := range []bool{false, true} {
		for phase := int32(0); phase < 2; phase++ {
			for bits := int32(0); bits < 8; bits++ {
				s := NewSolver()
				if xor {
					Bmcg_sat_solver_add_xor(s, 2, 0, 1, phase)
				} else {
					Bmcg_sat_solver_add_and(s, 2, 0, 1, 0, 0, phase)
				}
				as := []int32{int32(MkLit(0, bits&1 == 0)), int32(MkLit(1, bits&2 == 0)), int32(MkLit(2, bits&4 == 0))}
				x, y := bits&1 != 0, bits&2 != 0
				want := x && y
				if xor {
					want = x != y
				}
				want = want != (phase != 0)
				want = want == (bits&4 != 0)
				if got := Bmcg_sat_solver_solve(s, as, 3) == GLUCOSE_SAT; got != want {
					t.Fatal(xor, phase, bits, got, want)
				}
			}
		}
	}
}
