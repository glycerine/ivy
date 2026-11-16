package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/bsat"
	"os"
	"strings"
	"testing"
)

func TestDsdMaximumSupportedManagerFitsCaches(t *testing.T) {
	p := If_DsdManAlloc(16, 0)
	defer If_DsdManFree(p, 0)
	if If_DsdManVarNum(p) != 16 || If_DsdManObjNum(p) != 2 {
		t.Fatal("maximum-support manager geometry changed")
	}
}

func TestDsdFreeReleasesOwnedSATInstance(t *testing.T) {
	p := If_DsdManAlloc(4, 3)
	solver, ok := p.pSat.(*bsat.Solver)
	if !ok {
		t.Fatalf("missing original owned SAT solver: %T", p.pSat)
	}
	If_DsdManFree(p, 0)
	if solver.NumVars() != 0 || p.pSat != nil {
		t.Fatalf("freed DSD retained SAT state: %d vars, handle%T", solver.NumVars(), p.pSat)
	}
}

func TestDsdVerboseFreeWritesOriginalPrimeCaches(t *testing.T) {
	t.Chdir(t.TempDir())
	p := If_DsdManAlloc(5, 2)
	solver := p.pSat.(*bsat.Solver)
	If_DsdObjFindOrAdd(p, int32(IF_DSD_PRIME), []int32{2, 2, 2}, 3, []word{0x1717171717171717})
	If_DsdObjFindOrAdd(p, int32(IF_DSD_PRIME), []int32{2, 2, 2, 2}, 4, []word{0xD3C5D3C5D3C5D3C5})
	output := dsdReportCapture(t, func() { If_DsdManFree(p, 1) })
	for _, tc := range []struct {
		arity int
		want  string
	}{{3, "1717171717171717\n"}, {4, "D3C5D3C5D3C5D3C5\n"}, {5, ""}} {
		name := fmt.Sprintf("tt_dumpdsd%02d_%02d.txt", tc.arity, tc.arity)
		data, err := os.ReadFile(name)
		if err != nil {
			t.Errorf("missing cache %s: %v", name, err)
			continue
		}
		if string(data) != tc.want {
			t.Errorf("cache %s: %q, want%q", name, data, tc.want)
		}
		message := fmt.Sprintf("Dumped %d %d-var truth tables into file %q (0.00 MB).\n", len(strings.Fields(tc.want)), tc.arity, name)
		if !strings.Contains(output, message) {
			t.Errorf("missing cache report %q", message)
		}
	}
	if solver.NumVars() != 0 || p.pSat != nil || If_DsdManObjNum(p) != 0 {
		t.Fatal("verbose teardown retained owned state")
	}
}

func TestDsdVerboseFreeReleasesSATOnDumpFailure(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("tt_dumpdsd03_03.txt", 0700); err != nil {
		t.Fatal(err)
	}
	p := If_DsdManAlloc(3, 2)
	solver := p.pSat.(*bsat.Solver)
	var failure any
	dsdReportCapture(t, func() {
		defer func() { failure = recover() }()
		If_DsdManFree(p, 1)
	})
	if failure == nil {
		t.Fatal("expected the existing vector dump API's file-open panic")
	}
	if solver.NumVars() != 0 || p.pSat != nil || If_DsdManObjNum(p) != 0 {
		t.Fatal("dump failure leaked owned SAT or manager state")
	}
}
