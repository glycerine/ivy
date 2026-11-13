package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose"
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

// cecSat.c:818-861 and cecSatG.c:800-858 keep the SAT-ID map only in
// cone mode; a non-cone query clears the object IDs before returning.
func TestCec23SolveTwoConeOptionControlsSATIDLifetime(t *testing.T) {
	for _, cones := range []int32{0, 1} {
		t.Run("cec2", func(t *testing.T) {
			g := gia.Gia_ManStart(8)
			gia.Gia_ManAppendCi(g)
			gia.Gia_ManAppendCi(g)
			var opts Cec2_Par_t
			Cec2_SetDefaultParams(&opts)
			opts.fUseCones = cones
			p := Cec2_ManCreate(g, &opts)
			defer Cec2_ManDestroy(p)
			if got := Cec2_ManSolveTwo(p, 1, 2, 0); got != satoko.SATOKO_SAT {
				t.Fatalf("status %d", got)
			}
			if cones != 0 && Cec2_ManSolveTwo(p, 2, 1, 1) != satoko.SATOKO_SAT {
				t.Fatal("reversed repeated cone query failed")
			}
			for _, id := range []int32{1, 2} {
				v := Cec2_ObjSatId(p.pNew, gia.Gia_ManObj(p.pNew, id))
				if (v >= 0) != (cones != 0) {
					t.Fatalf("cones=%d id=%d sat=%d", cones, id, v)
				}
			}
		})
		t.Run("cec3", func(t *testing.T) {
			g := gia.Gia_ManStart(8)
			gia.Gia_ManAppendCi(g)
			gia.Gia_ManAppendCi(g)
			var opts Cec3_Par_t
			Cec3_SetDefaultParams(&opts)
			opts.fUseCones = cones
			p := Cec3_ManCreate(g, &opts)
			defer Cec3_ManDestroy(p)
			if got := Cec3_ManSolveTwo(p, 1, 2, 0); got != glucose.GLUCOSE_SAT {
				t.Fatalf("status %d", got)
			}
			if cones != 0 && Cec3_ManSolveTwo(p, 2, 1, 1) != glucose.GLUCOSE_SAT {
				t.Fatal("reversed repeated cone query failed")
			}
			for _, id := range []int32{1, 2} {
				v := Cec3_ObjSatId(p.pNew, gia.Gia_ManObj(p.pNew, id))
				if (v >= 0) != (cones != 0) {
					t.Fatalf("cones=%d id=%d sat=%d", cones, id, v)
				}
			}
		})
	}
}
