package gia

import "testing"

func mappingReviewMuxControlCone() (*Gia_Man_t, int32, int32) {
	p := Gia_ManStart(24)
	p.EnableMuxes()
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	control := Gia_ManAppendAnd(p, a, b)
	control = Gia_ManAppendAnd(p, control, c)
	control = Gia_ManAppendAnd(p, control, d)
	root := Gia_ManAppendMuxReal(p, control, a, b)
	Gia_ManAppendCo(p, root)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	id := root >> 1
	Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
	for _, value := range []int32{4, a >> 1, b >> 1, c >> 1, d >> 1, id} {
		Vec_IntPush(p.VMapping, value)
	}
	return p, control >> 1, id
}
func TestMappingChoiceLevelIncludesMuxControl(t *testing.T) {
	p, control, root := mappingReviewMuxControlCone()
	defer Gia_ManStop(p)
	if got := Gia_ManChoiceLevel(p); got != 4 {
		t.Fatalf("choice depth%d want4 including MUX control", got)
	}
	if Gia_ObjLevelId(p, control) != 3 || Gia_ObjLevelId(p, root) != 4 {
		t.Fatal("control/root levels missing")
	}
}
func TestMappingOverlapIncludesMuxControlCone(t *testing.T) {
	p, _, root := mappingReviewMuxControlCone()
	defer Gia_ManStop(p)
	if got := Gia_ManComputeOverlapOne(p, root); got != 4 {
		t.Errorf("traversal cone nodes%d want4", got)
	}
	label, visited := Vec_StrStart(int(p.nObjs)), Vec_IntAlloc(4)
	if got := Gia_ManComputeOverlap2One(p, root, label, visited); got != 4 {
		t.Errorf("label cone nodes%d want4", got)
	}
	for i := 0; i < int(p.nObjs); i++ {
		if Vec_StrEntry(label, i) != 0 {
			t.Fatalf("label%d was not cleared", i)
		}
	}
	if Gia_ManComputeOverlap(p) != 0 || Gia_ManComputeOverlap2(p) != 0 {
		t.Error("fully covered graph should have zero overlap")
	}
}

func TestMappingOverlapRequiresMapping(t *testing.T) {
	for _, compute := range []func(*Gia_Man_t) int32{Gia_ManComputeOverlap, Gia_ManComputeOverlap2} {
		p := Gia_ManStart(4)
		a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		Gia_ManAppendAnd(p, a, b)
		cexReviewMustPanic(t, func() { compute(p) })
		Gia_ManStop(p)
	}
}

func TestMappingTypedDependenciesAndTruth(t *testing.T) {
	for depth := 1; depth <= 8; depth++ {
		for phase := 0; phase < 8; phase++ {
			p := Gia_ManStart(32)
			p.EnableMuxes()
			var inputs [4]int32
			for i := range inputs {
				inputs[i] = Gia_ManAppendCi(p)
			}
			control := Gia_ManAppendAnd(p, inputs[0], inputs[1])
			for i := 1; i < depth; i++ {
				control = Gia_ManAppendAnd(p, control, inputs[(i+1)%4])
			}
			data := Gia_ManAppendXorReal(p, inputs[0]^int32(phase&1), inputs[1])
			other := Gia_ManAppendBuf(p, inputs[2]^int32(phase>>1&1))
			root := Gia_ManAppendMuxReal(p, control^int32(phase>>2&1), data, other)
			Gia_ManAppendCo(p, root)
			p.VMapping = Vec_IntStart(int(p.nObjs))
			id := root >> 1
			Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
			for _, v := range []int32{4, 1, 2, 3, 4, id} {
				Vec_IntPush(p.VMapping, v)
			}
			if got := Gia_ManChoiceLevel(p); got != int32(max(depth, 1)+1) {
				t.Fatalf("depth%d phase%d level%d", depth, phase, got)
			}
			if got := Gia_ManComputeOverlapOne(p, id); got != int32(depth+3) {
				t.Fatalf("typed cone%d want%d", got, depth+3)
			}
			label, visited := Vec_StrStart(int(p.nObjs)), Vec_IntAlloc(16)
			if got := Gia_ManComputeOverlap2One(p, id, label, visited); got != int32(depth+3) {
				t.Fatal("typed label cone mismatch")
			}
			if Gia_ManComputeOverlap(p) != 0 || Gia_ManComputeOverlap2(p) != 0 {
				t.Fatal("typed full coverage should have zero overlap")
			}
			for bits := 0; bits < 16; bits++ {
				var values [4]bool
				for i := range values {
					values[i] = bits>>i&1 != 0
				}
				choose := values[0] && values[1]
				for i := 1; i < depth; i++ {
					choose = choose && values[(i+1)%4]
				}
				choose = choose != (phase>>2&1 != 0)
				want := values[2] != (phase>>1&1 != 0)
				if choose {
					want = (values[0] != (phase&1 != 0)) != values[1]
				}
				got, err := p.Evaluate(values[:])
				if err != nil || got[0] != want {
					t.Fatalf("depth%d phase%d bits%d got%v want%v err%v", depth, phase, bits, got, want, err)
				}
			}
			Gia_ManStop(p)
		}
	}
}
