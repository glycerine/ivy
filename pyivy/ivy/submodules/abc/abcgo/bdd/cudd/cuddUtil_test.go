package cudd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"unsafe"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestCuddCountLeavesCountsDistinctTerminals(t *testing.T) {
	dd := NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)

	if got := Cudd_CountLeaves(dd.One); got != 1 {
		t.Fatalf("constant leaf count = %d, want 1", got)
	}
	f := core.Cudd_bddAnd(dd, x, y)
	if got := Cudd_CountLeaves(f); got != 1 {
		t.Fatalf("BDD terminal leaf count = %d, want 1", got)
	}
	if got := Cudd_CountLeaves(core.Cudd_Not(f)); got != 1 {
		t.Fatalf("complemented BDD terminal leaf count = %d, want 1", got)
	}
}

func TestCuddRandomMatchesPortableSequence(t *testing.T) {
	Cudd_Srandom(1)
	want := []int64{1902191391, 688417274, 489581952, 888303591, 991623880}
	for i, expected := range want {
		if got := Cudd_Random(); got != expected {
			t.Fatalf("Cudd_Random() #%d = %d, want %d", i, got, expected)
		}
	}

	Cudd_Srandom(0)
	zeroSeed := Cudd_Random()
	Cudd_Srandom(1)
	if got := Cudd_Random(); got != zeroSeed {
		t.Fatalf("seed 0 first value = %d, seed 1 first value = %d", zeroSeed, got)
	}
	Cudd_Srandom(-7)
	negativeSeed := Cudd_Random()
	Cudd_Srandom(7)
	if got := Cudd_Random(); got != negativeSeed {
		t.Fatalf("seed -7 first value = %d, seed 7 first value = %d", negativeSeed, got)
	}
}

func TestCuddDensityUsesManagerSizeForZeroNvars(t *testing.T) {
	dd := NewBDDManager(3)
	if got, want := Cudd_Density(dd, dd.One, 0), 8.0; got != want {
		t.Fatalf("constant-one density = %g, want %g", got, want)
	}
}

func TestCuddEstimateCofactorSimple(t *testing.T) {
	dd := NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x, y)

	for _, index := range []int32{0, 1} {
		if got := Cudd_EstimateCofactorSimple(f, index); got != 2 {
			t.Errorf("positive cofactor estimate for variable %d = %d, want 2", index, got)
		}
	}
	if got := Cudd_EstimateCofactorSimple(dd.One, 0); got != 1 {
		t.Errorf("constant cofactor estimate = %d, want 1", got)
	}
}

func TestCuddEstimateCofactorByPhase(t *testing.T) {
	dd := NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x, y)
	if got := Cudd_EstimateCofactor(dd, f, 0, 1); got != 2 {
		t.Errorf("positive cofactor estimate = %d, want 2", got)
	}
	if got := Cudd_EstimateCofactor(dd, f, 0, 0); got != 1 {
		t.Errorf("negative cofactor estimate = %d, want 1", got)
	}
}

func TestCuddSharingSizeCountsSharedNodesOnce(t *testing.T) {
	dd := NewBDDManager(1)
	x := core.Cudd_bddIthVar(dd, 0)
	notX := core.Cudd_Not(x)
	if got := Cudd_SharingSize([]*DdNode{x, notX}, 2); got != 2 {
		t.Fatalf("sharing size for a function and its complement = %d, want 2", got)
	}
}

func TestCuddCountPathCountsBranches(t *testing.T) {
	dd := NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x, y)
	if got := Cudd_CountPath(f); got != 3 {
		t.Fatalf("path count for x AND y = %g, want 3", got)
	}
	if got := Cudd_CountPath(dd.One); got != 1 {
		t.Fatalf("constant path count = %g, want 1", got)
	}
}

func TestCuddCountMintermHandlesComplementedEdges(t *testing.T) {
	dd := NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x, core.Cudd_Not(y))
	if got := Cudd_CountMinterm(dd, f, 2); got != 1 {
		t.Fatalf("minterm count for x AND !y = %g, want 1", got)
	}
	if got := Cudd_CountMinterm(dd, core.Cudd_Not(f), 2); got != 3 {
		t.Fatalf("minterm count for complement = %g, want 3", got)
	}
}

func TestCuddCubeConversionsAndComputeCube(t *testing.T) {
	dd := NewBDDManager(3)
	vars := []*DdNode{core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)}
	phase := []int32{1, 0}
	cube := Cudd_bddComputeCube(dd, vars, phase, 2)
	if cube == nil {
		t.Fatal("Cudd_bddComputeCube returned nil")
	}
	values := []int32{2, 2, 2}
	if Cudd_BddToCubeArray(dd, cube, values) != 1 {
		t.Fatal("Cudd_BddToCubeArray rejected a computed cube")
	}
	want := []int32{1, 0, 2}
	for i := range want {
		if values[i] != want[i] {
			t.Fatalf("cube values = %v, want %v", values, want)
		}
	}
	if restored := Cudd_CubeArrayToBdd(dd, values); restored != cube {
		t.Fatal("cube array conversion did not recover canonical BDD")
	}
}

func TestCuddSupportUtilities(t *testing.T) {
	dd := NewBDDManager(3)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 2)
	f := core.Cudd_bddAnd(dd, x, y)
	if got := Cudd_SupportIndex(dd, f); len(got) != 3 || got[0] != 1 || got[1] != 0 || got[2] != 1 {
		t.Fatalf("support index = %v, want [1 0 1]", got)
	}
	if got := Cudd_VectorSupportSize(dd, []*DdNode{f, x}, 2); got != 2 {
		t.Fatalf("vector support size = %d, want 2", got)
	}
}

func TestCuddClassifySupport(t *testing.T) {
	dd := NewBDDManager(3)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	z := core.Cudd_bddIthVar(dd, 2)
	f := core.Cudd_bddAnd(dd, x, y)
	g := core.Cudd_bddAnd(dd, x, z)
	var common, onlyF, onlyG *DdNode
	if Cudd_ClassifySupport(dd, f, g, &common, &onlyF, &onlyG) != 1 {
		t.Fatal("Cudd_ClassifySupport failed")
	}
	wantF := core.Cudd_bddAnd(dd, dd.One, y)
	wantG := core.Cudd_bddAnd(dd, dd.One, z)
	if common != x || onlyF != wantF || onlyG != wantG {
		t.Fatalf("classified cubes = (%p, %p, %p), want (%p, %p, %p)", common, onlyF, onlyG, x, wantF, wantG)
	}
}

func TestCuddPickArbitraryMintermsReturnsDistinctSatisfyingCubes(t *testing.T) {
	dd := NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddOr(dd, x, y)
	minterms := Cudd_bddPickArbitraryMinterms(dd, f, []*DdNode{x, y}, 2, 3)
	if len(minterms) != 3 {
		t.Fatalf("picked %d minterms, want 3", len(minterms))
	}
	seen := make(map[*DdNode]struct{})
	for _, minterm := range minterms {
		if _, duplicate := seen[minterm]; duplicate {
			t.Fatal("returned a duplicate minterm")
		}
		seen[minterm] = struct{}{}
		for mask := 0; mask < 4; mask++ {
			assignment := []bool{mask&1 != 0, mask&2 != 0}
			want, _ := dd.Evaluate(f, assignment)
			got, _ := dd.Evaluate(minterm, assignment)
			if got && !want {
				t.Fatalf("picked minterm accepts %v outside the function", assignment)
			}
		}
	}
}

func TestCuddMintermCountsIncludeSkippedVariables(t *testing.T) {
	dd := NewBDDManager(2)
	if got := Cudd_CountMinterm(dd, dd.One, 5); got != 32 {
		t.Fatalf("constant minterm count over five variables = %g, want 32", got)
	}
	var count EpDouble
	if status := Cudd_EpdCountMinterm(dd, dd.One, 1300, &count); status != 0 {
		t.Fatalf("extended count status = %d, want success", status)
	}
	if count.Value != 0.5 || count.Exponent != 1301 {
		t.Fatalf("extended count = %g * 2^%d, want 0.5 * 2^1301", count.Value, count.Exponent)
	}
}

func TestCuddCubeAndNodeGenerators(t *testing.T) {
	dd := NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddOr(dd, x, y)
	var cube *int32
	var value float64
	gen := Cudd_FirstCube(dd, f, &cube, &value)
	if gen == nil || Cudd_IsGenEmpty(gen) != 0 || value != 1 {
		t.Fatal("first cube generator did not produce a cube")
	}
	if got := unsafe.Slice(cube, 2); got[0] != 0 || got[1] != 1 {
		t.Fatalf("first path cube = %v, want [0 1]", got)
	}
	seen := 1
	for Cudd_NextCube(gen, &cube, &value) != 0 {
		seen++
		if got := unsafe.Slice(cube, 2); got[0] != 1 || got[1] != 2 {
			t.Fatalf("second path cube = %v, want [1 2]", got)
		}
	}
	if seen != 2 || Cudd_IsGenEmpty(gen) != 1 {
		t.Fatalf("cube generator yielded %d cubes, want 2", seen)
	}
	Cudd_GenFree(gen)

	var node *DdNode
	nodes := Cudd_FirstNode(dd, f, &node)
	if nodes == nil || node == nil {
		t.Fatal("node generator did not return its first node")
	}
	seen = 1
	for Cudd_NextNode(nodes, &node) != 0 {
		seen++
	}
	if seen != int(Cudd_DagSize(f)) {
		t.Fatalf("node generator yielded %d nodes, want %d", seen, Cudd_DagSize(f))
	}
}

func TestCuddPrintMintermPrintsDisjointPathCube(t *testing.T) {
	dd := NewBDDManager(3)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x, core.Cudd_Not(y))
	var output bytes.Buffer
	core.Cudd_SetOutput(dd, &output, nil)
	status := Cudd_PrintMinterm(dd, f)
	if status != 1 || strings.TrimSpace(output.String()) != "10-  1" {
		t.Fatalf("Cudd_PrintMinterm status/output = %d/%q, want 1/\"10-  1\"", status, output.String())
	}
}

type flushBuffer struct {
	bytes.Buffer
	flushes int
}

func (b *flushBuffer) Flush() error {
	b.flushes++
	return nil
}

func TestCuddPrintDebugUsesManagerOutputAndFlushes(t *testing.T) {
	dd := NewBDDManager(1)
	var output flushBuffer
	core.Cudd_SetOutput(dd, &output, nil)

	if got := Cudd_PrintDebug(dd, nil, 1, 0); got != 0 {
		t.Fatalf("NULL DD status = %d, want 0", got)
	}
	if got, want := output.String(), ": is the NULL DD\n"; got != want {
		t.Fatalf("NULL DD output = %q, want %q", got, want)
	}
	if output.flushes != 1 {
		t.Fatalf("NULL DD flush count = %d, want 1", output.flushes)
	}

	output.Reset()
	output.flushes = 0
	if got := Cudd_PrintDebug(dd, dd.LogicZero, 1, 1); got != 1 {
		t.Fatalf("zero DD status = %d, want 1", got)
	}
	if got, want := output.String(), ": is the zero DD\n"; got != want {
		t.Fatalf("zero DD output = %q, want %q", got, want)
	}
	if output.flushes != 1 {
		t.Fatalf("zero DD flush count = %d, want 1", output.flushes)
	}
}

func TestCuddPrintDebugNodeRecordsPreserveComplementAndEdges(t *testing.T) {
	dd := NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x, y)
	if f == nil {
		t.Fatal("failed to construct test BDD")
	}
	regular := Cudd_Regular(f)

	for _, tc := range []struct {
		name string
		root *DdNode
		mark byte
	}{
		{name: "regular", root: f, mark: ' '},
		{name: "complemented", root: Cudd_Not(f), mark: '!'},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			core.Cudd_SetOutput(dd, &output, nil)
			if got := cuddP(dd, tc.root); got != 1 {
				t.Fatalf("cuddP status = %d, want 1", got)
			}
			want := fmt.Sprintf("ID = %c0x%x\tindex = %d\tr = %d\t", tc.mark, regular.ID, regular.Index, regular.Ref)
			tRecord := ""
			tChild := Cudd_T(regular)
			if Cudd_IsConstant(Cudd_Regular(tChild)) {
				tRecord = fmt.Sprintf("T = %-9g\t", Cudd_Regular(tChild).Value)
			} else {
				tRecord = fmt.Sprintf("T = 0x%x\t", Cudd_Regular(tChild).ID)
			}
			eChild := Cudd_E(regular)
			var eRecord string
			if Cudd_IsConstant(Cudd_Regular(eChild)) {
				bang := " "
				if Cudd_IsComplement(eChild) {
					bang = "!"
				}
				eRecord = fmt.Sprintf("E = %c%-9g\n", bang[0], Cudd_Regular(eChild).Value)
			} else {
				bang := " "
				if Cudd_IsComplement(eChild) {
					bang = "!"
				}
				eRecord = fmt.Sprintf("E = %c0x%x\n", bang[0], Cudd_Regular(eChild).ID)
			}
			want += tRecord + eRecord
			if !strings.Contains(output.String(), want) {
				t.Fatalf("complete node record %q missing from output %q", want, output.String())
			}
			if !strings.HasSuffix(output.String(), "\n\n") {
				t.Fatalf("cuddP output lacks C trailing blank line: %q", output.String())
			}
		})
	}
}

func TestCuddCubeGeneratorMatchesSourceTraversalOrder(t *testing.T) {
	dd := NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddOr(dd, x, y)
	var cube *int32
	var value float64
	gen := Cudd_FirstCube(dd, f, &cube, &value)
	if gen == nil || cube == nil || value != 1 {
		t.Fatal("Cudd_FirstCube did not return first onset cube")
	}
	cubeValues := unsafe.Slice(cube, 2)
	got := [][2]int32{{cubeValues[0], cubeValues[1]}}
	for Cudd_NextCube(gen, &cube, &value) != 0 {
		cubeValues = unsafe.Slice(cube, 2)
		got = append(got, [2]int32{cubeValues[0], cubeValues[1]})
	}
	want := [][2]int32{{0, 1}, {1, 2}}
	if len(got) != len(want) {
		t.Fatalf("cube traversal = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("cube traversal = %v, want %v", got, want)
		}
	}
	Cudd_GenFree(gen)

	// The complemented function follows the same edge-phase rules and has
	// exactly the all-zero assignment as its onset.
	gen = Cudd_FirstCube(dd, core.Cudd_Not(f), &cube, &value)
	if gen == nil || cube == nil || value != 1 || unsafe.Slice(cube, 2)[0] != 0 || unsafe.Slice(cube, 2)[1] != 0 {
		t.Fatalf("complemented cube = %v, value %g; want [0 0], 1", cube, value)
	}
	if Cudd_NextCube(gen, &cube, &value) != 0 {
		t.Fatal("complemented function yielded an extra cube")
	}
	Cudd_GenFree(gen)
}

func TestCuddFirstCubeReturnsInitializedEmptyGenerator(t *testing.T) {
	dd := NewBDDManager(2)
	var cube *int32
	value := -1.0
	gen := Cudd_FirstCube(dd, dd.LogicZero, &cube, &value)
	if gen == nil || cube == nil || value != 0 || Cudd_IsGenEmpty(gen) != 1 {
		t.Fatalf("empty first cube = gen %v cube %p value %g status %d", gen != nil, cube, value, Cudd_IsGenEmpty(gen))
	}
	if Cudd_NextCube(gen, &cube, &value) != 0 {
		t.Fatal("empty function yielded a cube")
	}
	Cudd_GenFree(gen)
}
