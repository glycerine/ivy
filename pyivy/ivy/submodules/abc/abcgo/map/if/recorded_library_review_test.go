package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestRecordedLibraryCutPermHasWritableStorage(t *testing.T) {
	p := If_ManStart(&If_Par_t{NLutSize: 6, NCutsMax: 4, FUsePerm: 1})
	defer If_ManStop(p)
	var cut If_Cut_t
	If_CutSetup(p, &cut)
	cut.nLeaves = 2
	pins := If_CutPerm(&cut)
	if len(pins) < 2 {
		t.Fatalf("source cut has no writable pin storage: %d", len(pins))
	}
	pins[0], pins[1] = 1, 3
	again := If_CutPerm(&cut)
	if &pins[0] != &again[0] || again[0] != 1 || again[1] != 3 {
		t.Fatal("pin storage is not retained by cut")
	}
	var copy If_Cut_t
	If_CutCopy(p, &copy, &cut)
	If_CutPerm(&copy)[0] = 9
	if If_CutPerm(&cut)[0] != 1 {
		t.Fatal("copied pin storage aliases source")
	}
}

func TestRecordedLibraryBindingAndReentrantRegistration(t *testing.T) {
	old := recordedLibraryBackend()
	defer RegisterRecordedLibrary(old)
	p, cut, obj := &If_Man_t{}, &If_Cut_t{}, &If_Obj_t{}
	calls := 0
	RegisterRecordedLibrary(RecordedLibrary{DelayCost: func(got *If_Man_t, c *If_Cut_t, o *If_Obj_t) int32 {
		calls++
		if got != p || c != cut || o != obj {
			t.Fatal("backend arguments changed")
		}
		c.fUser, c.Cost = 1, 7
		// Invocations must not retain the registration lock.
		RegisterRecordedLibrary(RecordedLibrary{DelayCost: func(*If_Man_t, *If_Cut_t, *If_Obj_t) int32 { return 23 }})
		return 11
	}})
	if got := If_CutDelayRecCost3(p, cut, obj); got != 11 || calls != 1 || cut.fUser != 1 || cut.Cost != 7 {
		t.Fatal("cost callback result/mutations lost")
	}
	if got := If_CutDelayRecCost3(p, cut, obj); got != 23 {
		t.Fatal("new registration not used")
	}
	RegisterRecordedLibrary(RecordedLibrary{})
	defer func() {
		if got := recover(); got != "if: recorded-library delay backend is not registered" {
			t.Fatalf("missing backend panic%v", got)
		}
	}()
	If_CutDelayRecCost3(p, cut, obj)
}

func TestRecordedLibraryDefaultBuilderAndOverride(t *testing.T) {
	old := recordedLibraryBackend()
	defer RegisterRecordedLibrary(old)
	for _, override := range []bool{false, true} {
		p := fromIfReviewRecordedManager()
		if !override {
			p.pPars.SetRecordedGiaBuilder(nil)
		}
		calls := 0
		RegisterRecordedLibrary(RecordedLibrary{BuildGia: func(dst *Gia_Man_t, manager *If_Man_t, cut *If_Cut_t, leaves *Vec_Int_t, hash int32) int32 {
			calls++
			if override {
				t.Fatal("registered builder replaced the explicit override")
			}
			if manager != p || cut != If_ObjCutBest(If_ManObj(p, 3)) || Vec_IntSize(leaves) != 2 || hash != 0 {
				t.Fatal("reconstruction arguments changed")
			}
			return gia.Gia_ManAppendAnd(dst, Vec_IntEntry(leaves, 0), Vec_IntEntry(leaves, 1))
		}})
		result := Gia_ManFromIfAig(p)
		if calls != 1 && !override || calls != 0 && override {
			t.Fatal("wrong builder call count")
		}
		if Gia_ManCiNum(result) != 2 || Gia_ManCoNum(result) != 3 {
			t.Fatal("default builder changed interface")
		}
		Gia_ManStop(result)
		If_ManStop(p)
	}
}

func TestRecordedPinStorageDoesNotSlowOrdinaryTiming(t *testing.T) {
	p, cut, _ := ifTimeTestManager([]float32{2, 7, 3})
	defer If_ManStop(p)
	var delay float32
	allocations := testing.AllocsPerRun(20, func() {
		cut.ext = nil
		delay = If_CutDelay(p, nil, cut)
	})
	if delay != 8 || allocations != 0 {
		t.Fatalf("ordinary cut delay%f allocations%f", delay, allocations)
	}
}
