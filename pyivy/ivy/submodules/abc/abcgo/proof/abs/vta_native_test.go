package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestVtaRejectsNativeGatesWithoutChangingClasses(t *testing.T) {
	for _, mux := range []bool{false, true} {
		for _, internal := range []bool{false, true} {
			g := gia.Gia_ManStart(10)
			x := gia.Gia_ManAppendCi(g)
			y := gia.Gia_ManAppendCi(g)
			z := gia.Gia_ManAppendCi(g)
			var out int32
			if mux {
				g.EnableMuxes()
				out = gia.Gia_ManAppendMuxReal(g, x, y, z)
			} else {
				out = gia.Gia_ManAppendXorReal(g, x, y)
			}
			gia.Gia_ManAppendCo(g, out)
			seed := vec.Vec_IntAlloc(4)
			for _, value := range []int32{1, 3, 4, out >> 1} {
				vec.Vec_IntPush(seed, value)
			}
			gia.Gia_ManSetObjClasses(g, seed)
			stale := util.Abc_CexAlloc(0, 3, 1)
			util.Abc_InfoSetBit(stale.PData, 0)
			if mux {
				util.Abc_InfoSetBit(stale.PData, 1)
			}
			if err := g.SetSequentialCounterexample(stale); err != nil {
				t.Fatal(err)
			}
			var pars Abs_Par_t
			Abs_ParSetDefaults(&pars)
			pars.nFramesMax = 2
			var got int32
			if internal {
				got = Gia_VtaPerformInt(g, &pars)
			} else {
				got = Gia_VtaPerform(g, &pars)
			}
			if got != -1 || g.SequentialCounterexample() != nil || gia.Gia_ManObjClasses(g) != seed {
				t.Fatalf("mux=%v internal=%v result=%d or changed state", mux, internal, got)
			}
		}
	}
}
