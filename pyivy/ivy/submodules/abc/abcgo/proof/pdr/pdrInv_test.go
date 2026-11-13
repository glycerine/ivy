package pdr

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func pdrInvVector(values ...int32) *vec.Vec_Int_t {
	return vec.Vec_IntAllocArrayCopy(values, len(values))
}

func TestInvariantManagerCollectionAndFlopCounts(t *testing.T) {
	g := aig.NewGraph()
	for i := 0; i < 3; i++ {
		e := g.NewInput()
		g.AddOutput(e)
	}
	aig.Aig_ManSetRegNum(g, 3)
	frames := vec.Vec_VecStart(4)
	first := Pdr_SetCreate(pdrInvVector(0, 3), pdrInvVector())
	second := Pdr_SetCreate(pdrInvVector(3, 4), pdrInvVector())
	ignored := Pdr_SetCreate(pdrInvVector(1), pdrInvVector())
	ignored.nRefs = -1
	vec.Vec_VecPush(frames, 2, first)
	vec.Vec_VecPush(frames, 3, second)
	vec.Vec_VecPush(frames, 3, ignored)
	p := &Pdr_Man_t{pAig: g, vClauses: frames, vSolvers: make([]*bsat.Solver, 4)}
	if got := Pdr_ManFindInvariantStart(p); got != 1 {
		t.Fatalf("start=%d", got)
	}
	cubes := Pdr_ManCollectCubes(p, 2)
	if got := vec.Vec_PtrSize(cubes); got != 3 {
		t.Fatalf("cube count=%d", got)
	}
	if got := vec.Vec_IntArray(Pdr_ManCountFlops(p, cubes)); !reflect.DeepEqual(got, []int32{1, 2, 1}) {
		t.Fatalf("flop counts=%v", got)
	}
	if got := Pdr_ManCountVariables(p, 2); got != 3 {
		t.Fatalf("used variables=%d", got)
	}
}

func TestInvariantReductionReleasesMarkedCube(t *testing.T) {
	g := aig.NewGraph()
	_ = g.NewInput() // register output
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const1()) // the initially false register becomes true
	aig.Aig_ManSetRegNum(g, 1)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	p := Pdr_ManStart(g, &pars, nil)
	Pdr_ManCreateSolver(p, 0)
	Pdr_ManCreateSolver(p, 1)
	cube := Pdr_SetCreate(pdrInvVector(0), pdrInvVector())
	vec.Vec_VecPush(p.vClauses, 1, cube)
	packed := Pdr_ManDeriveInfinityClauses(p, 1)
	if cube.nRefs != -1 || !reflect.DeepEqual(vec.Vec_IntArray(packed), []int32{0, 1}) {
		t.Fatalf("noninductive cube not removed: refs=%d invariant=%v", cube.nRefs, vec.Vec_IntArray(packed))
	}
	Pdr_ManStop(p) // C decrements the sentinel when it releases the cube.
}

func TestInvariantCountsAndCover(t *testing.T) {
	// Two cubes, (flop0=1, flop2=0) and (flop2=1), among four flops.
	inv := pdrInvVector(2, 2, 0, 5, 1, 4, 4)
	counts := Pdr_InvCounts(inv)
	if got := vec.Vec_IntArray(counts); !reflect.DeepEqual(got, []int32{1, 0, 2, 0}) {
		t.Fatalf("counts=%v", got)
	}
	if got := Pdr_InvUsedFlopNum(inv); got != 2 {
		t.Fatalf("used flops=%d", got)
	}
	if got := vec.Vec_IntArray(Pdr_InvMap(counts)); !reflect.DeepEqual(got, []int32{0, 0, 1, 0}) {
		t.Fatalf("mapping=%v", got)
	}
	if got := string(vec.Vec_StrArray(Pdr_InvPrintStr(inv, counts))); got != "10 1\n-1 1\n\x00" {
		t.Fatalf("cover=%q", got)
	}
}

func TestInvariantEmptyAndMalformed(t *testing.T) {
	inv := pdrInvVector(0, 3)
	if got := Pdr_InvUsedFlopNum(inv); got != 0 {
		t.Fatalf("used flops=%d", got)
	}
	if got := string(vec.Vec_StrArray(Pdr_InvPrintStr(inv, Pdr_InvCounts(inv)))); got != "\x00" {
		t.Fatalf("empty cover=%q", got)
	}
	for _, malformed := range [][]int32{{}, {1, 2, 0, 1}, {1, 1, 4, 2}, {0, 1, 2}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted malformed invariant %v", malformed)
				}
			}()
			Pdr_InvCounts(pdrInvVector(malformed...))
		}()
	}
}

func TestInvariantDuplicateAIG(t *testing.T) {
	g := aig.NewGraph()
	x, q, r := g.NewInput(), g.NewInput(), g.NewInput()
	g.Name, g.Spec = "name", "spec"
	g.AddOutput(x)
	g.AddOutput(g.AND(x, q.Not()))
	g.AddOutput(g.NewXor(q, r))
	aig.Aig_ManSetRegNum(g, 2)
	cubes := vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(cubes, Pdr_SetCreate(pdrInvVector(0, 3), pdrInvVector()))
	vec.Vec_PtrPush(cubes, Pdr_SetCreate(pdrInvVector(), pdrInvVector()))
	dup := Pdr_ManDupAigWithClauses(g, cubes)
	if dup.Name != g.Name || dup.Spec != g.Spec || dup.RegisterCount() != 2 {
		t.Fatal("metadata not preserved")
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		got, err := dup.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{in[1] && !in[2], true, in[0] && !in[1], in[1] != in[2]}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("input %v: got %v, want %v", in, got, want)
		}
	}
}

func TestInvariantCheckGia(t *testing.T) {
	for _, test := range []struct {
		name            string
		nextOne, badOne bool
		want            int32
	}{
		{"inductive", false, false, 0}, {"transition fails", true, false, 1}, {"property fails", false, true, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := gia.Gia_ManStart(16)
			q := gia.Gia_ManAppendCi(g)
			bad := q
			if test.badOne {
				bad = 1
			}
			next := q
			if test.nextOne {
				next = 1
			}
			gia.Gia_ManAppendCo(g, bad)
			gia.Gia_ManAppendCo(g, next)
			gia.Gia_ManSetRegNum(g, 1)
			inv := pdrInvVector(1, 1, 0, 1)
			if got := Pdr_InvCheck(g, inv, 0); got != test.want {
				t.Fatalf("got %d, want %d", got, test.want)
			}
		})
	}
}

func TestInvariantCheckUnknownReturnsCountSoFarAsInC(t *testing.T) {
	g := gia.Gia_ManStart(16)
	input := gia.Gia_ManAppendCi(g)
	latch := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, input)
	gia.Gia_ManAppendCo(g, latch)
	gia.Gia_ManSetRegNum(g, 1)
	solver := pdrInvGiaSolver(g)
	if solver == nil {
		t.Fatal("invariant CNF unavailable")
	}
	solver.SetDeadline(time.Now().Add(-time.Second))
	if status := solver.Solve(bsat.Positive(1)); status != bsat.Unknown {
		t.Fatalf("expected an inconclusive query, got %v", status)
	}
	// With no cubes, C breaks out on an unknown coverage query and returns
	// the failures accumulated so far, which is zero in this case.
	if got := Pdr_InvCheck_int(g, pdrInvVector(0, 1), 0, solver, 0); got != 0 {
		t.Fatalf("got failures=%d, want source's zero", got)
	}
}

func TestInvariantMinimization(t *testing.T) {
	g := gia.Gia_ManStart(16)
	q := gia.Gia_ManAppendCi(g)
	r := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, q)
	gia.Gia_ManAppendCo(g, q)
	gia.Gia_ManAppendCo(g, r)
	gia.Gia_ManSetRegNum(g, 2)
	t.Run("duplicate clause", func(t *testing.T) {
		inv := pdrInvVector(2, 1, 0, 1, 0, 2)
		got := Pdr_InvMinimize(g, inv, 0)
		if got == nil || !reflect.DeepEqual(vec.Vec_IntArray(got), []int32{1, 1, 0, 2}) {
			t.Fatalf("minimized = %v", got)
		}
		if Pdr_InvCheck(g, got, 0) != 0 {
			t.Fatal("minimized invariant is invalid")
		}
	})
	t.Run("redundant literal", func(t *testing.T) {
		// q || r and q || !r together imply q; adding r to a third cube is redundant.
		inv := pdrInvVector(3, 1, 0, 1, 0, 2, 0, 2, 2)
		got := Pdr_InvMinimizeLits(g, inv, 0)
		if !slices.Contains(vec.Vec_IntArray(inv), int32(-1)) {
			t.Fatal("source leaves removed literals marked in the input")
		}
		if got == nil || Pdr_InvCheck(g, got, 0) != 0 {
			t.Fatalf("invalid minimized invariant: %v", got)
		}
	})
}

func TestInvariantDumpWithoutGlobalFrame(t *testing.T) {
	g := aig.NewGraph()
	for i := 0; i < 2; i++ {
		e := g.NewInput()
		g.AddOutput(e)
	}
	aig.Aig_ManSetRegNum(g, 2)
	frames := vec.Vec_VecStart(2)
	vec.Vec_VecPush(frames, 1, Pdr_SetCreate(pdrInvVector(0), pdrInvVector()))
	p := &Pdr_Man_t{pAig: g, pPars: &Pdr_Par_t{fUseSupp: 1}, vClauses: frames, vSolvers: make([]*bsat.Solver, 2)}
	path := filepath.Join(t.TempDir(), "inv.pla")
	Pdr_ManDumpClauses(p, path, 1)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{".i 1\n", ".o 1\n", ".p 1\n", "1 1\n", ".e\n"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("dump lacks %q: %s", want, data)
		}
	}
}
