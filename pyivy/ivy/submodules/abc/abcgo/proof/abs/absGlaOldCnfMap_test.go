package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestInterleavedInputsProbe(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n := gia.Gia_ManAppendAnd(g, a, b)
	c := gia.Gia_ManAppendCi(g)
	root := gia.Gia_ManAppendAnd(g, n, c)
	gia.Gia_ManAppendCo(g, root)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(root>>1), 1)
	gia.Gia_ManSetGateClasses(g, classes)
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.NFramesMax = 1
	if Gia_ManPerformGlaOld(g, &p, 0) != 0 {
		t.Fatal("expected SAT")
	}
}

func TestOldGlaCnfMappingPolarity(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n := gia.Gia_ManAppendAnd(g, a, b^1)
	c := gia.Gia_ManAppendCi(g)
	root := gia.Gia_ManAppendAnd(g, n, c)
	po := gia.Gia_ManAppendCo(g, root^1)
	data, err := glaOldCnfForGia(g)
	if err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 8; bits++ {
		v := make([]bool, gia.Gia_ManObjNum(g))
		v[a>>1] = bits&1 != 0
		v[b>>1] = bits&2 != 0
		v[c>>1] = bits&4 != 0
		v[n>>1] = v[a>>1] && !v[b>>1]
		v[root>>1] = v[n>>1] && v[c>>1]
		v[po>>1] = !v[root>>1]
		for _, clause := range data.Clauses {
			sat := false
			for _, lit := range clause {
				b := v[lit>>1]
				if lit&1 != 0 {
					b = !b
				}
				sat = sat || b
			}
			if !sat {
				t.Fatalf("bits=%d clause=%v", bits, clause)
			}
		}
	}
}
