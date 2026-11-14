package aigpartsat

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

func runMappedSource(t *testing.T, arguments ...string) string {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("native oracle uses GNU section collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	src := filepath.Clean(filepath.Join(dir, "../../../src"))
	binary := filepath.Join(t.TempDir(), "mapped-source")
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src, filepath.Join(dir, "test_vectors/mapped_incremental.c")}
	for _, folder := range []string{"aig/aig", "sat/cnf", "opt/dar", "bool/kit"} {
		files, err := filepath.Glob(filepath.Join(src, folder, "*.c"))
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range files {
			// kitPerm.c unconditionally uses a Windows-only integer type;
			// neither it nor the legacy demo is used by this call graph.
			if filepath.Base(name) != "kitPerm.c" && filepath.Base(name) != "aigTest.c" {
				args = append(args, name)
			}
		}
	}
	for _, name := range []string{"sat/bsat/satSolver.c", "sat/bsat/satMem.c", "sat/bsat/satStore.c", "sat/bsat/satUtil.c", "misc/util/utilSort.c", "misc/extra/extraUtilMisc.c", "misc/tim/timMan.c", "misc/mem/mem.c"} {
		args = append(args, filepath.Join(src, name))
	}
	args = append(args, "-Wl,--gc-sections", "-Wl,--wrap=Dar_ManRwsat", "-lm", "-o", binary)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile mapped source: %v: %s", err, out)
	}
	out, err := exec.CommandContext(ctx, binary, arguments...).CombinedOutput()
	if err != nil {
		t.Fatalf("run mapped source: %v: %s", err, out)
	}
	return string(out)
}

func TestMappedIncrementalClausesMatchOriginalC(t *testing.T) {
	want := runMappedSource(t)
	var got strings.Builder
	for align := 0; align < 2; align++ {
		g := aig.NewGraph()
		root := g.Const1()
		var pio []int32
		for i := 0; i < 6; i++ {
			input := g.NewInput()
			pio = append(pio, int32(input.NodeID()))
			root = g.AND(root, input)
		}
		g.AddOutput(root)
		pio = append(pio, int32(root.NodeID()))
		if align == 0 {
			aig.Aig_ManPartResetNodePolarity(g)
		}
		s := bsat.NewSolver()
		s.EnableClauseStore()
		variables := NewNodeVariables(int(aig.Aig_ManObjNumMax(g)))
		for stage := 0; stage < 2; stage++ {
			if Aig_ManAddNewCnfToSolver(s, g, variables, pio, nil, int32(align)) != 0 {
				t.Fatal("CNF unexpectedly inconsistent")
			}
			fmt.Fprintf(&got, "align %d stage %d vars %d map:", align, stage, s.NumVars())
			for _, v := range variables.IDs {
				fmt.Fprintf(&got, " %d", v)
			}
			fmt.Fprintln(&got)
		}
		store := s.ReleaseClauseStore()
		for _, clause := range store.Clauses() {
			fmt.Fprint(&got, "clause:")
			for _, l := range clause.Literals {
				fmt.Fprintf(&got, " %d", l)
			}
			fmt.Fprintln(&got)
		}
		bsat.Sto_ManFree(store)
		s.Close()
	}
	if got.String() != want {
		t.Fatalf("Go incremental mappings/clauses:\n%sC mappings/clauses:\n%s", got.String(), want)
	}
}

// Preserve an explicit demonstration of the upstream premature-SAT bug;
// source fidelity is not a reason to return a witness violating the circuit.
func TestSourceDriverEdgeCasesKeepCorrectResults(t *testing.T) {
	source := runMappedSource(t, "zero-budget")
	if !strings.Contains(source, "source status 0 witness valid 0") {
		t.Fatalf("source premature-SAT observation changed: %s", source)
	}
	if !strings.Contains(source, "source parts 1 output groups 8 first outputs 0") {
		t.Fatalf("source sparse-label observation changed: %s", source)
	}
	partitionGraph := aig.NewGraph()
	x, y := partitionGraph.NewInput(), partitionGraph.NewInput()
	root := partitionGraph.AND(x, y)
	partitionGraph.AddOutput(root)
	labels := make([]int32, aig.Aig_ManObjNumMax(partitionGraph))
	labels[root.NodeID()] = 7
	parts, _, outputs := aig.Aig_ManPartSplit(partitionGraph, labels)
	if len(parts) != 1 || len(outputs) != 1 || len(outputs[0]) != 1 {
		t.Fatal("sparse labels detached original outputs")
	}
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(g.AND(a.Not(), b))
	if status := Aig_ManPartitionedSat(g, 2, 1, 0, 0, 0, 0, 0); status != 1 {
		t.Fatalf("contradictory outputs: status %d, want UNSAT", status)
	}
}
