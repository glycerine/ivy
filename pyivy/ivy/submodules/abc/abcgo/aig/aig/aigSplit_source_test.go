package aig

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestSplitBddPipelineAgainstCompiledSource(t *testing.T) {
	defer Aig_ManRandom(1)
	var input, want strings.Builder
	want.WriteString("subset: 8 16 20 22\n")
	for table := 0; table < 256; table++ {
		g := splitThreeInputTruth(uint8(table))
		var function *cudd.DdNode
		dd := Aig_ManBuildPoBdd(g, &function)
		method, enabled := cudd.Cudd_ReorderingStatus(dd)
		if !enabled || method != cudd.CUDD_REORDER_SYMM_SIFT {
			t.Fatal("property BDD builder did not enable source symmetric sifting")
		}
		for selection := 0; selection < 16; selection++ {
			if selection == 8 {
				g.AddOutput(g.NewXor(g.inputs[0], g.inputs[2]))
				Aig_ManSetRegNum(g, 1)
			}
			fmt.Fprintf(&input, "%d %d\n", table, selection)
			var subset []*cudd.DdNode
			var indices []int
			for i := 2; i >= 0; i-- {
				if selection&(1<<i) != 0 {
					subset = append(subset, dd.Vars[i])
					indices = append(indices, i)
				}
			}
			cofactors := Aig_ManCofactorBdds(g, subset, dd, function)
			q := Aig_ManConvertBddsToAigs(g, dd, cofactors)
			if err := q.Check(); err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(&want, "%d %d %d %d:", table, selection, len(cofactors), len(q.outputs))
			var nonzero []uint8
			for code, cofactor := range cofactors {
				var truth uint8
				for mask := 0; mask < 8; mask++ {
					assignment := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
					got, err := dd.Evaluate(cofactor, assignment)
					expected := table&(1<<mask) != 0
					for i, index := range indices {
						expected = expected && (mask>>index)&1 == (code>>(len(indices)-1-i))&1
					}
					if err != nil || got != expected {
						t.Fatalf("table %d selection %d cofactor %d mask %d: got %t (%v), want %t", table, selection, code, mask, got, err, expected)
					}
					if got {
						truth |= 1 << mask
					}
				}
				fmt.Fprintf(&want, " %d", truth)
				if truth != 0 {
					nonzero = append(nonzero, truth)
				}
			}
			if q.nRegs != g.nRegs || len(q.outputs)-int(q.nRegs) != len(nonzero) {
				t.Fatalf("zero cofactor omission/order: table %d selection %d", table, selection)
			}
			want.WriteString(" | ")
			for i := range q.outputs {
				var truth uint8
				for mask := 0; mask < 8; mask++ {
					out, err := q.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
					if err != nil {
						t.Fatal(err)
					}
					if out[i] {
						truth |= 1 << mask
					}
				}
				expected := uint8(90) // The final LI is input[0] XOR state input[2].
				if i < len(nonzero) {
					expected = nonzero[i]
				}
				if truth != expected {
					t.Fatalf("MUX/state reconstruction: table %d selection %d output %d = %d, want %d", table, selection, i, truth, expected)
				}
				fmt.Fprintf(&want, "%d ", truth)
			}
			want.WriteByte('\n')
			for _, cofactor := range cofactors {
				cudd.Cudd_RecursiveDeref(dd, cofactor)
			}
		}
		cudd.Cudd_RecursiveDeref(dd, function)
		if count := cudd.Cudd_CheckZeroRef(dd); count != 0 {
			t.Fatalf("table %d leaked %d BDD roots", table, count)
		}
		cudd.Cudd_Quit(dd)
	}
	for table := 1; table < 255; table++ {
		for _, selection := range []int{16, 24} {
			g := splitThreeInputTruth(uint8(table))
			if selection&8 != 0 {
				g.AddOutput(g.NewXor(g.inputs[0], g.inputs[2]))
				Aig_ManSetRegNum(g, 1)
			}
			fmt.Fprintf(&input, "%d %d\n", table, selection)
			Aig_ManRandom(1)
			q := Aig_ManSplit(g, 2, 0)
			if q == nil || q.nRegs != g.nRegs {
				t.Fatalf("full split failed for table %d selection %d", table, selection)
			}
			fmt.Fprintf(&want, "split %d %d %d: ", table, selection, len(q.outputs))
			truths := make([]uint8, len(q.outputs))
			for mask := 0; mask < 8; mask++ {
				out, err := q.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
				if err != nil {
					t.Fatal(err)
				}
				any := false
				for i, bit := range out {
					if bit {
						truths[i] |= 1 << mask
					}
					if i < len(out)-int(q.nRegs) {
						any = any || bit
					}
				}
				if any != (table&(1<<mask) != 0) {
					t.Fatalf("full split changed property: table %d mask %d", table, mask)
				}
			}
			if q.nRegs != 0 && truths[len(truths)-1] != 90 {
				t.Fatal("full split changed latch input")
			}
			for _, truth := range truths {
				fmt.Fprintf(&want, "%d ", truth)
			}
			want.WriteByte('\n')
		}
	}
	_, file, _, _ := runtime.Caller(0)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(file), "../../../src/bdd/cudd/*.c"))
	if err != nil || len(files) == 0 {
		t.Fatalf("CUDD sources: %v", err)
	}
	sources := []string{"aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c",
		"aigFanout.c", "aigUtil.c", "aigCheck.c", "aigDup.c", "aigDfs.c",
		"../../misc/tim/timMan.c", "../../misc/mem/mem.c", "../../bdd/mtr/mtrBasic.c", "../../bdd/mtr/mtrGroup.c",
		"../../misc/st/st.c", "../../misc/st/stmm.c", "../../misc/extra/extraUtilUtil.c", "../../bdd/extrab/extraBddMisc.c"}
	for _, file := range files {
		if filepath.Base(file) != "testcudd.c" {
			sources = append(sources, "../../bdd/cudd/"+filepath.Base(file))
		}
	}
	got := runAigSourceFixtureOptions(t, "split_bdd_compare.c", input.String(), 45*time.Second, []string{"-DABC_USE_CUDD", "-lm"}, sources...)
	if got != want.String() {
		gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
		for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
			if gotLines[i] != wantLines[i] {
				t.Fatalf("line %d: C %q, Go %q", i, gotLines[i], wantLines[i])
			}
		}
		t.Fatalf("C has %d trace lines, Go has %d", len(gotLines), len(wantLines))
	}
}

func splitThreeInputTruth(table uint8) *Graph {
	g := NewGraph()
	inputs := []Edge{g.NewInput(), g.NewInput(), g.NewInput()}
	root := g.Const0()
	for mask := 0; mask < 8; mask++ {
		if table&(1<<mask) == 0 {
			continue
		}
		term := g.Const1()
		for i, input := range inputs {
			if mask&(1<<i) == 0 {
				input = input.Not()
			}
			term = g.AND(term, input)
		}
		root = g.OR(root, term)
	}
	g.AddOutput(root)
	return g
}
