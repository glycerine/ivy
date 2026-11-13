package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// ABC's VTA manager asserts nObjs <= mask even though IDs end at nObjs-1.
// At powers of two this precondition rejects a representable object layout.
func TestVtaSourceRejectsPowerOfTwoObjectCount(t *testing.T) {
	g := gia.Gia_ManStart(6)
	q := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, q)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 1)
	if gia.Gia_ManObjNum(g) != 4 {
		t.Fatal("fixture must have four objects")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected the C source's object-mask assertion")
		}
	}()
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	pars.nFramesMax = 1
	Gia_VtaPerform(g, &pars)
}

// C VTA leaves the preexisting sequential witness untouched when it returns
// without finding a new counterexample, including its trivial-UNSAT branch.
func TestVtaSourceRetainsExistingWitnessWithoutNewCounterexample(t *testing.T) {
	for _, constant := range []bool{false, true} {
		g := gia.Gia_ManStart(8)
		gia.Gia_ManAppendCi(g)
		q := gia.Gia_ManAppendCi(g)
		out := q
		if constant {
			out = 0
		}
		gia.Gia_ManAppendCo(g, out)
		gia.Gia_ManAppendCo(g, 0)
		gia.Gia_ManSetRegNum(g, 1)
		stale := util.Abc_CexAlloc(1, 1, 1)
		if err := g.SetSequentialCounterexample(stale); err != nil {
			t.Fatal(err)
		}
		saved := g.SequentialCounterexample()
		var pars Abs_Par_t
		Abs_ParSetDefaults(&pars)
		pars.nFramesMax = 1
		pars.nRatioMin = 0
		got := Gia_VtaPerform(g, &pars)
		if got == 0 || g.SequentialCounterexample() != saved {
			t.Fatalf("constant=%v result=%d: source witness ownership changed", constant, got)
		}
	}
}

func TestVtaSourceConversionRejectsPowerOfTwoObjectCount(t *testing.T) {
	g := gia.Gia_ManStart(6)
	q := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, q)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 1)
	if gia.Gia_ManObjNum(g) != 4 {
		t.Fatal("fixture must have four objects")
	}
	gla := vec.Vec_IntStart(4)
	vec.Vec_IntWriteEntry(gla, 0, 1)
	packed := vec.Vec_IntAlloc(4)
	for _, v := range []int32{1, 3, 4, 0} {
		vec.Vec_IntPush(packed, v)
	}
	for _, convert := range []func(){func() { Gia_VtaConvertFromGla(g, gla, 1) }, func() { Gia_VtaConvertToGla(g, packed) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected C source object-mask assertion")
				}
			}()
			convert()
		}()
	}
}
