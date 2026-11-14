package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaNativeGateWitness(t *testing.T) {
	for _, mux := range []bool{false, true} {
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
		var pars Abs_Par_t
		Abs_ParSetDefaults(&pars)
		pars.NFramesMax = 1
		pars.NRatioMin = 0
		pars.NRatioMin2 = 0
		result := Gia_ManPerformGla(g, &pars)
		if result != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
			t.Fatalf("mux=%v result=%d witness=%+v", mux, result, g.SequentialCounterexample())
		}
	}
}

func TestGlaNativeSequentialWitness(t *testing.T) {
	for _, mux := range []bool{false, true} {
		for _, seeded := range []bool{false, true} {
			g := gia.Gia_ManStart(16)
			x := gia.Gia_ManAppendCi(g)
			q := gia.Gia_ManAppendCi(g)
			var special int32
			if mux {
				g.EnableMuxes()
				special = gia.Gia_ManAppendMuxReal(g, q, x, 0)
			} else {
				special = gia.Gia_ManAppendXorReal(g, q, x)
			}
			out := gia.Gia_ManAppendAnd2(g, special, q)
			gia.Gia_ManAppendCo(g, out)
			gia.Gia_ManAppendCo(g, 1)
			gia.Gia_ManSetRegNum(g, 1)
			if seeded {
				classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
				vec.Vec_IntWriteEntry(classes, 0, 1)
				vec.Vec_IntWriteEntry(classes, int(out>>1), 1)
				vec.Vec_IntWriteEntry(classes, int(q>>1), 1)
				gia.Gia_ManSetGateClasses(g, classes)
			}
			var pars Abs_Par_t
			Abs_ParSetDefaults(&pars)
			pars.NFramesMax = 4
			pars.NRatioMin = 0
			pars.NRatioMin2 = 0
			got := Gia_ManPerformGla(g, &pars)
			cex := g.SequentialCounterexample()
			if got != 0 || cex == nil || cex.IFrame != 1 || gia.Gia_ManVerifyCex(g, cex, 0) != 1 {
				t.Fatalf("mux=%v seeded=%v result=%d cex=%+v", mux, seeded, got, cex)
			}
		}
	}
}

func TestGlaNativeGateClassProjection(t *testing.T) {
	for _, mux := range []bool{false, true} {
		g := gia.Gia_ManStart(16)
		a := gia.Gia_ManAppendCi(g)
		b := gia.Gia_ManAppendCi(g)
		c := gia.Gia_ManAppendCi(g)
		var special int32
		if mux {
			g.EnableMuxes()
			special = gia.Gia_ManAppendMuxReal(g, c, a, b)
		} else {
			special = gia.Gia_ManAppendXorReal(g, a, b)
		}
		buffer := gia.Gia_ManAppendBuf(g, special)
		out := gia.Gia_ManAppendAnd(g, special, buffer^1)
		gia.Gia_ManAppendCo(g, out)
		var pars Abs_Par_t
		Abs_ParSetDefaults(&pars)
		pars.NFramesMax = 1
		pars.NRatioMin = 0
		pars.NRatioMin2 = 0
		if got := Gia_ManPerformGla(g, &pars); got != -1 {
			t.Fatalf("mux=%v result=%d", mux, got)
		}
		classes := gia.Gia_ManGateClasses(g)
		if classes == nil || vec.Vec_IntSize(classes) != int(gia.Gia_ManObjNum(g)) {
			t.Fatal("missing original gate map")
		}
		abstract := Gia_ManDupAbsGates(g, classes)
		n := len(abstract.InputLiterals())
		if n > 8 {
			t.Fatal("unexpected abstraction inputs")
		}
		for bits := 0; bits < 1<<n; bits++ {
			inputs := make([]bool, n)
			for i := range inputs {
				inputs[i] = bits&(1<<i) != 0
			}
			got, err := abstract.Evaluate(inputs)
			if err != nil || got[0] {
				t.Fatalf("mux=%v inputs=%v got=%v err=%v", mux, inputs, got, err)
			}
		}
	}
}

func TestGlaLowerTransfersVtaFrameClasses(t *testing.T) {
	g := gia.Gia_ManStart(12)
	x := gia.Gia_ManAppendCi(g)
	q := gia.Gia_ManAppendCi(g)
	special := gia.Gia_ManAppendXorReal(g, x, q)
	gia.Gia_ManAppendCo(g, special)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 1)
	gateClasses := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	for _, id := range []int32{0, q >> 1, special >> 1} {
		vec.Vec_IntWriteEntry(gateClasses, int(id), 1)
	}
	original := Gia_VtaConvertFromGla(g, gateClasses, 2)
	gia.Gia_ManSetObjClasses(g, original)
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	called := false
	result := glaPerformLowered(g, &pars, true, func(lowered *gia.Gia_Man_t) int32 {
		called = true
		mapped := gia.Gia_ManObjClasses(lowered)
		if mapped == nil || vec.Vec_IntEntry(mapped, 0) != 2 {
			t.Fatal("VTA frame map was not transferred")
		}
		frames := Gia_VtaAbsToFrames(mapped)
		bits := util.Abc_Base2Log(uint32(gia.Gia_ManObjNum(lowered)))
		mask := uint32(1)<<uint(bits) - 1
		for frame := 0; frame < vec.Vec_PtrSize(frames); frame++ {
			seenTimes := make(map[uint32]bool)
			for _, entry := range vec.Vec_IntArray(vec.Vec_PtrEntry(frames, frame).(*vec.Vec_Int_t)) {
				id := int32(uint32(entry) & mask)
				if gia.Gia_ObjIsAnd(gia.Gia_ManObj(lowered, id)) != 0 {
					seenTimes[uint32(entry)>>uint(bits)] = true
				}
			}
			for time := 0; time <= frame; time++ {
				if !seenTimes[uint32(time)] {
					t.Fatalf("frame %d lost lowered XOR cone at time %d", frame, time)
				}
			}
		}
		return -1
	})
	if !called || result != -1 || gia.Gia_ManObjClasses(g) != nil {
		t.Fatal("VTA classes were not consumed")
	}
	gia.Gia_ManSetObjClasses(g, original)
	glaPerformLowered(g, &pars, false, func(*gia.Gia_Man_t) int32 { return -1 })
	if gia.Gia_ManObjClasses(g) != original {
		t.Fatal("non-VTA run changed object classes")
	}
}

func TestOldGlaNativeVtaSeedIntegration(t *testing.T) {
	g := gia.Gia_ManStart(12)
	x := gia.Gia_ManAppendCi(g)
	q := gia.Gia_ManAppendCi(g)
	special := gia.Gia_ManAppendXorReal(g, x, q)
	out := gia.Gia_ManAppendAnd(g, special, q)
	gia.Gia_ManAppendCo(g, out)
	gia.Gia_ManAppendCo(g, 1)
	gia.Gia_ManSetRegNum(g, 1)
	gateClasses := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	for _, id := range []int32{0, q >> 1, special >> 1, out >> 1} {
		vec.Vec_IntWriteEntry(gateClasses, int(id), 1)
	}
	gia.Gia_ManSetObjClasses(g, Gia_VtaConvertFromGla(g, gateClasses, 1))
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	pars.NFramesStart = 2
	pars.NFramesMax = 4
	pars.NRatioMin = 0
	pars.NRatioMin2 = 0
	result := Gia_ManPerformGlaOld(g, &pars, 1)
	if result != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
		t.Fatalf("result=%d witness=%+v", result, g.SequentialCounterexample())
	}
	if gia.Gia_ManObjClasses(g) != nil {
		t.Fatal("VTA seed was not consumed")
	}
}

func TestOldGlaLoweringPreservesVtaSeedWhenVtaSkipped(t *testing.T) {
	for _, constant := range []bool{false, true} {
		g := gia.Gia_ManStart(10)
		x := gia.Gia_ManAppendCi(g)
		y := gia.Gia_ManAppendCi(g)
		_ = gia.Gia_ManAppendXorReal(g, x, y)
		out := x
		if constant {
			out = 0
		}
		gia.Gia_ManAppendCo(g, out)
		seed := vec.Vec_IntAlloc(4)
		for _, v := range []int32{1, 3, 4, 0} {
			vec.Vec_IntPush(seed, v)
		}
		gia.Gia_ManSetObjClasses(g, seed)
		var pars Abs_Par_t
		Abs_ParSetDefaults(&pars)
		pars.NFramesMax = 2
		pars.NFramesStart = 1
		result := Gia_ManPerformGlaOld(g, &pars, 1)
		if constant && result != 1 || !constant && result != 0 {
			t.Fatalf("constant=%v result=%d", constant, result)
		}
		if gia.Gia_ManObjClasses(g) != seed {
			t.Fatalf("constant=%v VTA seed cleared despite early return", constant)
		}
	}
}

func TestGlaNativeSeedDoesNotAbstractSimplifiedPrimaryInput(t *testing.T) {
	for _, old := range []bool{false, true} {
		g := gia.Gia_ManStart(10)
		a := gia.Gia_ManAppendCi(g)
		b := gia.Gia_ManAppendCi(g)
		special := gia.Gia_ManAppendXorReal(g, a, 0)
		out := gia.Gia_ManAppendAnd(g, special, b)
		gia.Gia_ManAppendCo(g, out)
		classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
		for _, id := range []int32{0, special >> 1, out >> 1} {
			vec.Vec_IntWriteEntry(classes, int(id), 1)
		}
		gia.Gia_ManSetGateClasses(g, classes)
		var pars Abs_Par_t
		Abs_ParSetDefaults(&pars)
		pars.NFramesMax = 1
		pars.NRatioMin = 0
		pars.NRatioMin2 = 0
		var result int32
		if old {
			result = Gia_ManPerformGlaOld(g, &pars, 0)
		} else {
			result = Gia_ManPerformGla(g, &pars)
		}
		if result != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
			t.Fatalf("old=%v result=%d witness=%+v", old, result, g.SequentialCounterexample())
		}
	}
}

func TestGlaNativeProjectionRetainsSimplifiedGateSemantics(t *testing.T) {
	for _, old := range []bool{false, true} {
		g := gia.Gia_ManStart(10)
		a := gia.Gia_ManAppendCi(g)
		special := gia.Gia_ManAppendXorReal(g, a, 0)
		out := gia.Gia_ManAppendAnd(g, special, a^1)
		gia.Gia_ManAppendCo(g, out)
		classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
		for _, id := range []int32{0, special >> 1, out >> 1} {
			vec.Vec_IntWriteEntry(classes, int(id), 1)
		}
		gia.Gia_ManSetGateClasses(g, classes)
		var pars Abs_Par_t
		Abs_ParSetDefaults(&pars)
		pars.NFramesMax = 1
		pars.NRatioMin = 0
		pars.NRatioMin2 = 0
		var result int32
		if old {
			result = Gia_ManPerformGlaOld(g, &pars, 0)
		} else {
			result = Gia_ManPerformGla(g, &pars)
		}
		if result == 0 {
			t.Fatalf("old=%v produced spurious witness", old)
		}
		mapped := gia.Gia_ManGateClasses(g)
		if mapped == nil {
			t.Fatalf("old=%v lost gate classes", old)
		}
		abstract := Gia_ManDupAbsGates(g, mapped)
		n := len(abstract.InputLiterals())
		if n > 8 {
			t.Fatal("unexpected abstract inputs")
		}
		for bits := 0; bits < 1<<n; bits++ {
			values := make([]bool, n)
			for i := range values {
				values[i] = bits&(1<<i) != 0
			}
			got, err := abstract.Evaluate(values)
			if err != nil || got[0] {
				t.Fatalf("old=%v inputs=%v output=%v err=%v", old, values, got, err)
			}
		}
	}
}

func TestGlaNativeProjectionRetainsRegisterAlias(t *testing.T) {
	for _, old := range []bool{false, true} {
		g := gia.Gia_ManStart(10)
		x := gia.Gia_ManAppendCi(g)
		q := gia.Gia_ManAppendCi(g)
		special := gia.Gia_ManAppendXorReal(g, q, 0)
		out := gia.Gia_ManAppendAnd(g, special, q^1)
		gia.Gia_ManAppendCo(g, out)
		gia.Gia_ManAppendCo(g, x)
		gia.Gia_ManSetRegNum(g, 1)
		classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
		for _, id := range []int32{0, special >> 1, out >> 1} {
			vec.Vec_IntWriteEntry(classes, int(id), 1)
		}
		gia.Gia_ManSetGateClasses(g, classes)
		var pars Abs_Par_t
		Abs_ParSetDefaults(&pars)
		pars.NFramesMax = 1
		pars.NRatioMin = 0
		pars.NRatioMin2 = 0
		var result int32
		if old {
			result = Gia_ManPerformGlaOld(g, &pars, 0)
		} else {
			result = Gia_ManPerformGla(g, &pars)
		}
		if result == 0 {
			t.Fatalf("old=%v returned spurious SAT", old)
		}
		mapped := gia.Gia_ManGateClasses(g)
		if mapped == nil {
			t.Fatalf("old=%v lost gate classes", old)
		}
		abstract := Gia_ManDupAbsGates(g, mapped)
		n := len(abstract.InputLiterals())
		if n > 8 {
			t.Fatal("unexpected abstract inputs")
		}
		for bits := 0; bits < 1<<n; bits++ {
			values := make([]bool, n)
			for i := range values {
				values[i] = bits&(1<<i) != 0
			}
			got, err := abstract.Evaluate(values)
			if err != nil || got[0] {
				t.Fatalf("old=%v inputs=%v output=%v err=%v", old, values, got, err)
			}
		}
	}
}

func TestGlaNativeProjectionRetainsAndAlias(t *testing.T) {
	for _, old := range []bool{false, true} {
		g := gia.Gia_ManStart(12)
		a := gia.Gia_ManAppendCi(g)
		b := gia.Gia_ManAppendCi(g)
		shared := gia.Gia_ManAppendAnd(g, a, b)
		special := gia.Gia_ManAppendXorReal(g, shared, 0)
		out := gia.Gia_ManAppendAnd(g, special, shared^1)
		gia.Gia_ManAppendCo(g, out)
		classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
		for _, id := range []int32{0, special >> 1, out >> 1} {
			vec.Vec_IntWriteEntry(classes, int(id), 1)
		}
		gia.Gia_ManSetGateClasses(g, classes)
		var pars Abs_Par_t
		Abs_ParSetDefaults(&pars)
		pars.NFramesMax = 1
		pars.NRatioMin = 0
		pars.NRatioMin2 = 0
		var result int32
		if old {
			result = Gia_ManPerformGlaOld(g, &pars, 0)
		} else {
			result = Gia_ManPerformGla(g, &pars)
		}
		if result == 0 {
			t.Fatalf("old=%v returned spurious SAT", old)
		}
		mapped := gia.Gia_ManGateClasses(g)
		if mapped == nil {
			t.Fatalf("old=%v lost gate classes", old)
		}
		abstract := Gia_ManDupAbsGates(g, mapped)
		n := len(abstract.InputLiterals())
		if n > 8 {
			t.Fatal("unexpected abstract inputs")
		}
		for bits := 0; bits < 1<<n; bits++ {
			values := make([]bool, n)
			for i := range values {
				values[i] = bits&(1<<i) != 0
			}
			got, err := abstract.Evaluate(values)
			if err != nil || got[0] {
				t.Fatalf("old=%v inputs=%v output=%v err=%v", old, values, got, err)
			}
		}
	}
}
