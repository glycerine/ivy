package abci

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestAbcNtkTestTimCreateArrivalAndRequiredVectors(t *testing.T) {
	arrivals := Abc_NtkTestCreateArrivals(12)
	required := Abc_NtkTestCreateRequired(3)
	defer vec.Vec_FltFree(arrivals)
	defer vec.Vec_FltFree(required)
	if got := vec.Vec_FltArray(arrivals); len(got) != 12 || got[0] != 0 || got[9] != 9 || got[10] != 0 || got[11] != 1 {
		t.Fatalf("unexpected arrival sequence: %v", got)
	}
	if got := vec.Vec_FltArray(required); len(got) != 3 || got[0] != 100 || got[1] != 101 || got[2] != 102 {
		t.Fatalf("unexpected required sequence: %v", got)
	}
}

func TestAbcNtkTestTimSynthesisPreservesTimingAndHoleGraph(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	root := gia.Gia_ManAppendAnd(p, a, b)
	gia.Gia_ManAppendCo(p, root)
	extra := gia.Gia_ManStart(2)
	input := gia.Gia_ManAppendCi(extra)
	gia.Gia_ManAppendCo(extra, input)
	manager := tim.Tim_ManStart(2, 1)
	gia.Gia_ManSetBoxData(p, extra, manager)
	defer func() {
		gia.Gia_ManStopTimingData(p)
		gia.Gia_ManStop(p)
	}()
	out := Abc_NtkTestTimPerformSynthesis(p, 0)
	if out == nil {
		t.Fatal("synthesis returned nil")
	}
	defer func() {
		gia.Gia_ManStopTimingData(out)
		gia.Gia_ManStop(out)
	}()
	gotManager, gotExtra := gia.Gia_ManBoxData(out)
	if gotManager == manager || gotExtra == extra {
		t.Fatal("synthesis must duplicate timing and extra graph ownership")
	}
	if _, ok := gotManager.(*tim.Tim_Man_t); !ok || gotExtra == nil {
		t.Fatalf("synthesis lost timing attachments: manager=%T extra=%p", gotManager, gotExtra)
	}
	if gia.Gia_ManCompare(p, out) == 0 {
		t.Fatal("synthesis changed the Boolean function")
	}
}
