package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestOldGlaBufferTriggersSourceExportAssertion(t *testing.T) {
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
	pars.NFramesMax = 1
	defer assertSimpleExportPanic(t)
	Gia_ManPerformGlaOld(g, &pars, 0)
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
