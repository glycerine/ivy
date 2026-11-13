package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"path/filepath"
	"testing"
)

func TestGlaDumpWritesAbstractedAiger(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	y := gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, gate)
	for _, lit := range []int32{x, y, gate} {
		gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(g, lit>>1), 1)
	}
	pars := &Abs_Par_t{fDumpVabs: 1, pFileVabs: filepath.Join(t.TempDir(), "abstract.aig")}
	p := &Ga2_Man_t{pGia: g, pPars: pars, vAbs: vec.Vec_IntAlloc(2)}
	vec.Vec_IntPush(p.vAbs, 0)
	vec.Vec_IntPush(p.vAbs, gate>>1)
	Ga2_GlaDumpAbsracted(p, 0)
	out, err := gia.Gia_AigerRead(pars.pFileVabs, 0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range []bool{false, true} {
		got, e := out.Evaluate([]bool{b, true})
		if e != nil || got[0] != b {
			t.Fatalf("input %v: %v, %v", b, got, e)
		}
	}
}

func TestGlaInitialDumpRestoresGateClasses(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	y := gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, gate)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	gia.Gia_ManSetGateClasses(g, classes)
	pars := &Abs_Par_t{pFileVabs: filepath.Join(t.TempDir(), "initial.aig")}
	p := &Ga2_Man_t{pGia: g, pPars: pars}
	Ga2_ManDumpInitialMap(p)
	if gia.Gia_ManGateClasses(g) != classes || vec.Vec_IntEntry(classes, int(gate>>1)) != 0 {
		t.Fatal("original gate classes not restored")
	}
	if _, err := gia.Gia_AigerRead(pars.pFileVabs, 0, 0, 1); err != nil {
		t.Fatal(err)
	}
}
