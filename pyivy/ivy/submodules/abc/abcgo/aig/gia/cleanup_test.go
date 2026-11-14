package gia

import "testing"

func TestCombinationalCleanupRetainsBuffersAndChoiceCones(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	shared := Gia_ManAppendAnd(p, a, b)
	unused := Gia_ManAppendAnd(p, a, c)
	buffered := Gia_ManAppendAnd(p, b, c)
	Gia_ManAppendBuf(p, buffered)
	Gia_ManAppendCo(p, shared)
	p.pSibls = make([]int32, p.nObjs)
	p.pSibls[buffered>>1] = unused >> 1
	cleaned := Gia_ManCleanup(p)
	// An unobserved buffer and its logic are explicit cleanup roots, as in C.
	if cleaned.nBufs != 1 || Gia_ManAndNum(cleaned) != 4 {
		t.Fatal("buffer/choice closure", Gia_ManAndNum(cleaned), cleaned.nBufs)
	}
	for _, obj := range p.pObjs {
		if obj.fMark0 != 0 {
			t.Fatal("cleanup did not clear marks")
		}
	}
}
func TestSequentialCleanupRemovesOnlyUnreachableRegisters(t *testing.T) {
	p := Gia_ManStart(8)
	pi, first, dead := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	next := Gia_ManAppendXorReal(p, pi, first)
	Gia_ManAppendCo(p, first)
	Gia_ManAppendCo(p, next)
	Gia_ManAppendCo(p, dead)
	Gia_ManSetRegNum(p, 2)
	p.VNamesIn = Vec_PtrAlloc(3)
	for _, name := range []string{"pi", "live", "dead"} {
		Vec_PtrPush(p.VNamesIn, name)
	}
	cleaned := Gia_ManSeqCleanup(p)
	if cleaned.nRegs != 1 || Gia_ManPiNum(cleaned) != 1 || Gia_ManPoNum(cleaned) != 1 || len(cleaned.InputNames()) != 2 || cleaned.InputNames()[1] != "live" {
		t.Fatal("register interface/names", cleaned.nRegs, cleaned.InputNames())
	}
	if p.pObjs[dead>>1].Value != ^uint32(0) {
		t.Fatal("removed object missing sentinel")
	}
	for mask := 0; mask < 4; mask++ {
		piValue, state := mask&1 != 0, mask&2 != 0
		got, err := cleaned.Evaluate([]bool{piValue, state})
		if err != nil || got[0] != state || got[1] != (piValue != state) {
			t.Fatal("cleaned transition", mask, got, err)
		}
	}
}
func TestSequentialCleanupNoOutputsAndUnreachableBuffer(t *testing.T) {
	p := Gia_ManStart(4)
	state := Gia_ManAppendCi(p)
	Gia_ManAppendBuf(p, state)
	Gia_ManAppendCo(p, state)
	Gia_ManSetRegNum(p, 1)
	cleaned := Gia_ManSeqCleanup(p)
	if cleaned.nRegs != 0 || Gia_ManCoNum(cleaned) != 0 || Gia_ManCiNum(cleaned) != 0 || cleaned.nObjs != 1 {
		t.Fatal("empty observable system cleanup")
	}
}
func TestCleanupRealMuxControlAndDroppedOutputConstraints(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	control := Gia_ManAppendAnd(p, a, b)
	p.EnableMuxes()
	mux := Gia_ManAppendMuxReal(p, control, a, c)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, mux)
	p.nConstrs = 1
	cleaned := Gia_ManCleanupOutputs(p, 1)
	if cleaned.nConstrs != 1 || Gia_ManCoNum(cleaned) != 1 || cleaned.nMuxes != 1 {
		t.Fatal("constraint cleanup metadata")
	}
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		before, err := p.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		after, err := cleaned.Evaluate(input)
		if err != nil || before[1] != after[0] {
			t.Fatal("cleaned MUX", mask, before, after, err)
		}
	}
}

func TestCleanupRetainsEquivalenceClassAndOwnsMetadata(t *testing.T) {
	p := Gia_ManStart(6)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	first, second := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, first)
	p.StartRepresentatives()
	Gia_ObjSetRepr(p, second>>1, first>>1)
	p.SetNext(first>>1, second>>1)
	cleaned := Gia_ManCleanup(p)
	if Gia_ManAndNum(cleaned) != 2 || Gia_ObjRepr(cleaned, second>>1) != first>>1 || cleaned.PNexts[first>>1] != second>>1 {
		t.Fatal("cleanup lost equivalent cone")
	}
	p.ClearRepresentatives()
	if Gia_ObjRepr(cleaned, second>>1) != first>>1 {
		t.Fatal("class metadata aliases source")
	}
}
