package gia

import "testing"

func frameTestGraph() *Gia_Man_t {
	p := Gia_ManStart(16)
	pi := Gia_ManAppendCi(p)
	ro := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, ro)
	Gia_ManAppendCo(p, pi)
	Gia_ManSetRegNum(p, 1)
	return p
}

func TestGiaManUnrollRejectsInvalidFrameCounts(t *testing.T) {
	for _, frames := range []int32{-1, 1<<30} {
		if got := Gia_ManUnrollStart(frameTestGraph(), &Gia_ParFra_t{nFrames: frames}); got != nil {
			t.Fatalf("unroll start with %d frames = %v, want nil", frames, got)
		}
	}
}

func TestGiaManUnrollIncrementalLifecycle(t *testing.T) {
	src := frameTestGraph()
	params := &Gia_ParFra_t{nFrames: 2, fSaveLastLit: 1}
	raw := Gia_ManUnrollStart(src, params)
	if raw == nil {
		t.Fatal("unroll start returned nil")
	}
	manager := raw.(*Gia_ManUnr_t)
	if got := Gia_ManCiNum(manager.pNew); got != 1 {
		t.Fatalf("initial CI count = %d, want symbolic register only", got)
	}
	first, ok := Gia_ManUnrollAdd(raw, 1).(*Gia_Man_t)
	if !ok || first == nil {
		t.Fatal("first incremental add did not return a GIA")
	}
	if got := Gia_ManCiNum(first); got != 2 {
		t.Fatalf("after frame 1 CI count = %d, want 2", got)
	}
	if got, want := Gia_ManUnrollLastLit(raw), Gia_Obj2Lit(first, Gia_ManPi(first, 0)); got != want {
		t.Fatalf("frame 1 property literal = %d, want initial RO %d", got, want)
	}
	second, ok := Gia_ManUnrollAdd(raw, 2).(*Gia_Man_t)
	if !ok || second != first {
		t.Fatal("second incremental add must return the same live graph")
	}
	if got := Gia_ManCiNum(second); got != 3 {
		t.Fatalf("after frame 2 CI count = %d, want 3", got)
	}
	if got, want := Gia_ManUnrollLastLit(raw), Gia_Obj2Lit(second, Gia_ManPi(second, 1)); got != want {
		t.Fatalf("frame 2 property literal = %d, want previous frame PI %d", got, want)
	}
	objectsBeforeStop := Gia_ManObjNum(second)
	Gia_ManUnrollStop(raw)
	if got := Gia_ManObjNum(second); got != objectsBeforeStop {
		t.Fatalf("stop changed returned graph object count: %d -> %d", objectsBeforeStop, got)
	}
	if got := Gia_ManUnrollLastLit(raw); got != -1 {
		t.Fatalf("last literal after stop = %d, want invalid sentinel", got)
	}
	Gia_ManAppendCo(second, 0)
}

func TestGiaManUnrollFixedFramesAndInitialization(t *testing.T) {
	src := frameTestGraph()
	params := &Gia_ParFra_t{nFrames: 2}
	raw := Gia_ManUnrollStart(src, params)
	manager := raw.(*Gia_ManUnr_t)
	if got := Gia_ManCiNum(manager.pNew); got != 3 {
		t.Fatalf("fixed-input initialization CI count = %d, want 2 frame PIs plus 1 state", got)
	}
	result, ok := Gia_ManUnrollAdd(raw, 2).(*Gia_Man_t)
	if !ok || result == nil || Gia_ManPoNum(result) != 2 {
		t.Fatalf("fixed unroll outputs = %v, want two frame properties", result)
	}
	Gia_ManHashStop(result)
	Gia_ManUnrollStop(raw)

	initialized := &Gia_ParFra_t{nFrames: 2, fInit: 1}
	result = Gia_ManFrames(src, initialized)
	if result == nil || Gia_ManRegNum(result) != 0 || Gia_ManPoNum(result) != 2 || Gia_ManCiNum(result) != 2 {
		t.Fatalf("initialized frames have CIs/POs/regs %d/%d/%d, want 2/2/0", Gia_ManCiNum(result), Gia_ManPoNum(result), Gia_ManRegNum(result))
	}
}

func TestGiaManFrames2KeepsStateInterface(t *testing.T) {
	result := Gia_ManFrames2(frameTestGraph(), &Gia_ParFra_t{nFrames: 3})
	if result == nil {
		t.Fatal("frame builder returned nil")
	}
	if got := Gia_ManRegNum(result); got != 1 {
		t.Fatalf("register count = %d, want 1", got)
	}
	if got := Gia_ManPoNum(result); got != 3 {
		t.Fatalf("property outputs = %d, want three frame outputs", got)
	}
	if got := Gia_ManCiNum(result); got != 4 {
		t.Fatalf("input count = %d, want three PIs and one symbolic initial state", got)
	}
}
