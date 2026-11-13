package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestOldGlaLocalCutsComplementAndConstant(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b^1), gia.Gia_ManAppendAnd(g, a, b^1)
	root := gia.Gia_ManAppendAnd(g, x, y)
	z := gia.Gia_ManAppendAnd(g, a, 0)
	gia.Gia_ManAppendCo(g, root)
	data, err := glaOldCnfForGia(g)
	if err != nil {
		t.Fatal(err)
	}
	if data.Obj2Count[z>>1] != 3 {
		t.Fatal("constant-input node lost")
	}
	for bits := 0; bits < 4; bits++ {
		v := make([]bool, gia.Gia_ManObjNum(g))
		v[a>>1] = bits&1 != 0
		v[b>>1] = bits&2 != 0
		v[x>>1] = v[a>>1] && !v[b>>1]
		v[y>>1] = v[x>>1]
		v[root>>1] = v[x>>1] && v[y>>1]
		v[gia.Gia_ObjId(g, gia.Gia_ManPo(g, 0))] = v[root>>1]
		for _, clause := range data.Clauses {
			sat := false
			for _, lit := range clause {
				value := v[lit>>1]
				if lit&1 != 0 {
					value = !value
				}
				sat = sat || value
			}
			if !sat {
				t.Fatalf("bits=%d clause=%v", bits, clause)
			}
		}
	}
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	for _, id := range []int32{x >> 1, y >> 1, root >> 1} {
		vec.Vec_IntWriteEntry(classes, int(id), 1)
	}
	gia.Gia_ManSetGateClasses(g, classes)
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	p := Gla_ManStart(g, &pars)
	defer Gla_ManStop(p)
	Gia_GlaAddTimeFrame(p, 0)
	if p.pSat.Solve(bsat.Literal(Gla_ManGetOutLit(p, 0))) != bsat.Satisfiable {
		t.Fatal("complemented AND should be satisfiable")
	}
}
