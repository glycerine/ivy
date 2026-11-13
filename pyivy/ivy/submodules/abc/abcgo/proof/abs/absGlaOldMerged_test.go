package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestOldGlaPreservesDistinctDuplicateNodes(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, b)
	root := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, root)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(root>>1), 1)
	gia.Gia_ManSetGateClasses(g, classes)
	data, err := glaOldCnfForGia(g)
	if err != nil {
		t.Fatal(err)
	}
	if data.Mapping[x>>1] == 0 || data.Mapping[y>>1] == 0 || data.Mapping[root>>1] == 0 {
		t.Fatal("distinct nodes lost from fallback")
	}
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	p := Gla_ManStart(g, &pars)
	defer Gla_ManStop(p)
	// Refine the two independent abstraction boundaries to recover the full graph.
	for _, id := range []int32{x >> 1, y >> 1} {
		wid := gia.Gia_ManObj(g, id).Value
		if wid == ^uint32(0) {
			t.Fatalf("missing node %d", id)
		}
		gid := p.pObj2Obj[wid]
		obj := Gla_ManObj(p, int32(gid))
		if obj.fAbs == 0 {
			obj.fAbs = 1
			vec.Vec_IntPush(p.vAbs, int32(gid))
		}
	}
	Gia_GlaAddTimeFrame(p, 0)
	out := Gla_ManGetOutLit(p, 0)
	for bits := 0; bits < 4; bits++ {
		assumptions := []bsat.Literal{bsat.Literal(out)}
		for i, id := range []int32{a >> 1, b >> 1} {
			wid := gia.Gia_ManObj(g, id).Value
			v := Gla_ManGetVar(p, int32(p.pObj2Obj[wid]), 0)
			lit := 2 * v
			if bits&(1<<i) == 0 {
				lit++
			}
			assumptions = append(assumptions, bsat.Literal(lit))
		}
		status := p.pSat.Solve(assumptions...)
		if (status == bsat.Satisfiable) != (bits == 3) {
			t.Fatalf("bits=%d status=%v", bits, status)
		}
	}
}

func TestOldGlaPreservesBuffer(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	buf := gia.Gia_ManAppendBuf(g, a^1)
	gia.Gia_ManAppendCo(g, buf)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(buf>>1), 1)
	gia.Gia_ManSetGateClasses(g, classes)
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	pars.nFramesMax = 1
	if result := Gia_ManPerformGlaOld(g, &pars, 0); result != 0 {
		t.Fatalf("result=%d", result)
	}
	if cex := g.SequentialCounterexample(); cex == nil || gia.Gia_ManVerifyCex(g, cex, 0) == 0 {
		t.Fatal("invalid buffer witness")
	}
}
func TestOldGlaRejectsUnloweredSpecialGates(t *testing.T) {
	for _, mux := range []bool{false, true} {
		g := gia.Gia_ManStart(10)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		var out int32
		if mux {
			g.EnableMuxes()
			out = gia.Gia_ManAppendMuxReal(g, a, b, c)
		} else {
			out = gia.Gia_ManAppendXorReal(g, a, b)
		}
		gia.Gia_ManAppendCo(g, out)
		if _, err := glaOldCnfForGia(g); err == nil {
			t.Fatalf("mux=%t accepted unlowered special gate", mux)
		}
	}
}
