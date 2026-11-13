package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestGlaPositiveRootAssignmentMatchesSource(t *testing.T) {
	g := gia.Gia_ManStart(12)
	gia.Gia_ManAppendCi(g)
	q, _ := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	root := gia.Gia_ManAppendAnd(g, q^1, 1)
	gia.Gia_ManAppendCo(g, root)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 2)
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.fUseSimple = 1
	p.fAddLayer = 1
	p.nFramesMax = 1
	p.nRatioMin = 0
	p.nRatioMin2 = 0
	p.nFramesNoChangeLim = 100
	result := Gia_ManPerformGla(g, &p)
	if result != -1 || g.SequentialCounterexample() != nil {
		t.Fatalf("source returned unknown without witness; Go result=%d witness=%v", result, g.SequentialCounterexample())
	}
	if binary := os.Getenv("ABS_GLA_ROOT_ORACLE"); binary != "" {
		out, err := exec.Command(binary).CombinedOutput()
		if err != nil || !strings.Contains(string(out), "SOURCE_STATUS=-1 CEX=0") {
			t.Fatalf("C oracle err=%v output=%s", err, out)
		}
	}
}
func TestGlaPositiveRootAssignmentReachableAfterRefinement(t *testing.T) {
	g := gia.Gia_ManStart(12)
	gia.Gia_ManAppendCi(g)
	q, _ := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	root := gia.Gia_ManAppendAnd(g, q^1, 1)
	gia.Gia_ManAppendCo(g, root)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 2)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(root>>1), 1)
	gia.Gia_ManSetGateClasses(g, classes)
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	pars.fUseSimple = 1
	pars.fAddLayer = 1
	pars.iFrame = -1
	p := Ga2_ManStart(g, &pars)
	defer Ga2_ManStop(p)
	Ga2_ManRestart(p)
	pars.iFrame = 0
	p.vId2Lit = append(p.vId2Lit, vec.Vec_IntAlloc(0))
	vec.Vec_IntFillExtra(Ga2_MapFrameMap(p, 0), vec.Vec_IntSize(p.vValues), -1)
	Ga2_ManAddAbsClauses(p, 0)
	lit := Ga2_ObjFindLit(p, gia.Gia_ManObj(g, root>>1), 0)
	for p.pSat.NumVars() < p.nSatVars {
		p.pSat.NewVariable()
	}
	if p.pSat.Solve(bsat.Literal(lit)) != bsat.Satisfiable {
		t.Fatal("initial abstraction not SAT")
	}
	add := Ga2_ManRefine(p)
	if lit != 3 || !reflect.DeepEqual(vec.Vec_IntArray(add), []int32{int32(q >> 1)}) {
		t.Fatalf("lit=%d refinements=%v", lit, vec.Vec_IntArray(add))
	}
	p.pSat.Simplify()
	p.pSat.Bookmark()
	Ga2_ManAddToAbs(p, add)
	value, assigned := p.pSat.RootValue(lit >> 1)
	if !assigned || value || lit&1 != 1 {
		t.Fatalf("expected root-true complemented literal: assigned=%v value=%v lit=%d", assigned, value, lit)
	}
}

func TestGlaConstantOneSourceOracle(t *testing.T) {
	binary := os.Getenv("ABS_GLA_CONST_ORACLE")
	if binary == "" {
		t.Skip("set ABS_GLA_CONST_ORACLE to run the compiled C fixture")
	}
	out, err := exec.Command(binary).CombinedOutput()
	if err == nil || !strings.Contains(string(out), "Assertion `Lit > 1' failed") {
		t.Fatalf("C expected literal assertion; err=%v output=%s", err, out)
	}
}
