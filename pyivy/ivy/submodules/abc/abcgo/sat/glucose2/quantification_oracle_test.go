package glucose2

import (
	"github.com/glycerine/abcgo/aig/aig"
	"strings"
	"testing"
)

// The oracle enumerates quantified variables directly and does not use the
// CNF or SOP routines under test.
func TestQuantificationAllTwoInputFunctions(t *testing.T) {
	for table := 0; table < 16; table++ {
		for keepMask := 0; keepMask < 4; keepMask++ {
			g := aig.NewGraph()
			x, y := g.NewInput(), g.NewInput()
			root := g.Const0()
			for bits := 0; bits < 4; bits++ {
				if table&(1<<bits) == 0 {
					continue
				}
				a, b := x, y
				if bits&1 == 0 {
					a = a.Not()
				}
				if bits&2 == 0 {
					b = b.Not()
				}
				root = g.OR(root, g.AND(a, b))
			}
			keep := func(_ any, i int32) int32 {
				if keepMask&(1<<i) != 0 {
					return 1
				}
				return 0
			}
			solvers := [2]*SimpSolver{NewABCSolver(), NewABCSolver()}
			result := bmcg2_sat_solver_quantify(solvers, g, int32(root), 1, keep, nil, nil)
			if result < 0 {
				t.Fatalf("table=%04b keep=%02b: interrupted", table, keepMask)
			}
			g.AddOutput(aig.Edge(result))
			for assignment := 0; assignment < 4; assignment++ {
				want := false
				for q := 0; q < 4; q++ {
					if ((q^assignment)&keepMask) == 0 && table&(1<<q) != 0 {
						want = true
					}
				}
				got, err := g.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0})
				if err != nil || got[0] != want {
					t.Fatalf("table=%04b keep=%02b assignment=%02b got=%v want=%v err=%v", table, keepMask, assignment, got, want, err)
				}
			}
		}
	}
}

func TestQuantificationAllThreeInputFunctions(t *testing.T) {
	for table := 0; table < 256; table++ {
		for keepMask := 0; keepMask < 8; keepMask++ {
			g := aig.NewGraph()
			ins := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
			root := g.Const0()
			for bits := 0; bits < 8; bits++ {
				if table&(1<<bits) == 0 {
					continue
				}
				cube := g.Const1()
				for i, input := range ins {
					if bits&(1<<i) == 0 {
						input = input.Not()
					}
					cube = g.AND(cube, input)
				}
				root = g.OR(root, cube)
			}
			keep := func(_ any, i int32) int32 {
				if keepMask&(1<<i) != 0 {
					return 1
				}
				return 0
			}
			result := bmcg2_sat_solver_quantify([2]*SimpSolver{NewABCSolver(), NewABCSolver()}, g, int32(root), 1, keep, nil, nil)
			if result < 0 {
				t.Fatalf("table=%02x keep=%03b: interrupted", table, keepMask)
			}
			g.AddOutput(aig.Edge(result))
			for assignment := 0; assignment < 8; assignment++ {
				want := false
				for q := 0; q < 8; q++ {
					if (q^assignment)&keepMask == 0 && table&(1<<q) != 0 {
						want = true
					}
				}
				got, err := g.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
				if err != nil || got[0] != want {
					t.Fatalf("table=%02x keep=%03b assignment=%03b got=%v want=%v err=%v", table, keepMask, assignment, got, want, err)
				}
			}
		}
	}
}

func TestEquivalenceAndOverlapAgainstTruthTables(t *testing.T) {
	for left := 0; left < 16; left++ {
		for right := 0; right < 16; right++ {
			g := aig.NewGraph()
			ins := []aig.Edge{g.NewInput(), g.NewInput()}
			build := func(table int) aig.Edge {
				root := g.Const0()
				for bits := 0; bits < 4; bits++ {
					if table&(1<<bits) == 0 {
						continue
					}
					a, b := ins[0], ins[1]
					if bits&1 == 0 {
						a = a.Not()
					}
					if bits&2 == 0 {
						b = b.Not()
					}
					root = g.OR(root, g.AND(a, b))
				}
				return root
			}
			a, b := build(left), build(right)
			for _, equiv := range []int32{0, 1} {
				got := bmcg2_sat_solver_equiv_overlap_check(NewABCSolver(), g, int32(a), int32(b), equiv)
				want := int32(0)
				if equiv != 0 && left == right || equiv == 0 && left&right != 0 {
					want = 1
				}
				if got != want {
					t.Fatalf("left=%04b right=%04b equiv=%d got=%d want=%d", left, right, equiv, got, want)
				}
			}
		}
	}
}

func TestSOPCubesAgainstThreeInputTruthTables(t *testing.T) {
	for table := 0; table < 256; table++ {
		g := aig.NewGraph()
		ins := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		root := g.Const0()
		for bits := 0; bits < 8; bits++ {
			if table&(1<<bits) == 0 {
				continue
			}
			cube := g.Const1()
			for i, input := range ins {
				if bits&(1<<i) == 0 {
					input = input.Not()
				}
				cube = g.AND(cube, input)
			}
			root = g.OR(root, cube)
		}
		g.AddOutput(root)
		sop := bmcg2_sat_solver_sop(g, 0)
		for assignment := 0; assignment < 8; assignment++ {
			actual := false
			for _, line := range strings.Split(strings.TrimSpace(sop), "\n") {
				fields := strings.Fields(line)
				if len(fields) == 0 {
					continue
				}
				if len(fields) == 1 && (fields[0] == "0" || fields[0] == "1") {
					if fields[0] == "1" {
						actual = true
					}
					continue
				}
				if len(fields) != 2 || len(fields[0]) != 3 || fields[1] != "1" {
					t.Fatalf("bad SOP %q", sop)
				}
				match := true
				for i, ch := range fields[0] {
					if ch == '0' && assignment&(1<<i) != 0 || ch == '1' && assignment&(1<<i) == 0 {
						match = false
					}
				}
				if match {
					actual = true
				}
			}
			if actual != (table&(1<<assignment) != 0) {
				t.Fatalf("table=%02x assignment=%03b sop=%q", table, assignment, sop)
			}
		}
	}
}
