package gia

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestTasSolveAllThreeInputFunctions(t *testing.T) {
	for table := 0; table < 256; table++ {
		g := Gia_ManStart(128)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		Gia_ManHashStart(g)
		root := int32(0)
		for assignment := 0; assignment < 8; assignment++ {
			if table&(1<<assignment) == 0 {
				continue
			}
			x, y, z := a, b, c
			if assignment&1 == 0 {
				x ^= 1
			}
			if assignment&2 == 0 {
				y ^= 1
			}
			if assignment&4 == 0 {
				z ^= 1
			}
			root = Gia_ManHashOr(g, root, Gia_ManHashAnd(g, Gia_ManHashAnd(g, x, y), z))
		}
		Gia_ManAppendCo(g, root)
		Gia_ManCreateRefs(g)
		Gia_ManFillValue(g)
		p := Tas_ManAlloc(g, 10000)
		for _, negated := range []bool{false, true} {
			literal := root
			expectedSAT := table != 0
			if negated {
				literal ^= 1
				expectedSAT = table != 255
			}
			got := Tas_ManSolve(p, Gia_ObjFromLit(g, literal), nil)
			if (got == 0) != expectedSAT {
				t.Fatalf("truth=%08b negated=%v got=%d", table, negated, got)
			}
			if got == 0 {
				values := make([]bool, 3)
				for _, lit := range vec.Vec_IntArray(Tas_ReadModel(p)) {
					values[lit>>1] = lit&1 == 0
				}
				result, err := g.Evaluate(values)
				if err != nil || result[0] == negated {
					t.Fatalf("truth=%08b model=%v result=%v err=%v", table, values, result, err)
				}
			}
		}
		if Tas_ManSolve(p, Gia_ObjFromLit(g, a), Gia_ObjFromLit(g, a^1)) != 1 {
			t.Fatal("opposite assumptions were accepted")
		}
		Tas_ManStop(p)
		Gia_ManStop(g)
	}
}
