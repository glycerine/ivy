package ssc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// sscClass.c Ssc_GiaClassesInit asserts p->pReprs == NULL. In particular,
// a sweep must not overwrite representative information owned by its caller.
func TestClassesRejectExistingRepresentatives(t *testing.T) {
	subject := gia.Gia_ManStart(8)
	input := gia.Gia_ManAppendCi(subject)
	gia.Gia_ManAppendCo(subject, input)
	subject.StartRepresentatives()
	gia.Gia_ObjSetRepr(subject, input>>1, 0)
	defer func() {
		if recover() == nil {
			t.Error("accepted pre-existing representative storage")
		}
		if got := gia.Gia_ObjRepr(subject, input>>1); got != 0 {
			t.Errorf("representative overwritten: %d", got)
		}
	}()
	Ssc_GiaClassesInit(subject, []bool{false})
}

func TestSweepRejectsExistingRepresentatives(t *testing.T) {
	subject := gia.Gia_ManStart(8)
	input := gia.Gia_ManAppendCi(subject)
	gia.Gia_ManAppendCo(subject, input)
	subject.StartRepresentatives()
	gia.Gia_ObjSetRepr(subject, input>>1, 0)
	care := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(care)
	gia.Gia_ManAppendCo(care, 0) // no violated care constraints
	defer func() {
		if recover() == nil {
			t.Error("sweep accepted existing representatives")
		}
		if got := gia.Gia_ObjRepr(subject, input>>1); got != 0 {
			t.Errorf("representative overwritten: %d", got)
		}
	}()
	Ssc_PerformSweepingInt(subject, care, nil)
}
