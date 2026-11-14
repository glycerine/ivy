package gia

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestSeqModelVecAccessorsTransferOwnership(t *testing.T) {
	manager := Gia_ManStart(1)
	defer Gia_ManStop(manager)
	first := vec.Vec_PtrAlloc(0)
	second := vec.Vec_PtrAlloc(0)
	defer vec.Vec_PtrFree(first)
	defer vec.Vec_PtrFree(second)

	Gia_ManSetSeqModelVec(manager, first)
	if got := Gia_ManReadSeqModelVec(manager); got != first {
		t.Fatal("read accessor did not return the manager's raw vector")
	}

	Gia_ManSetSeqModelVec(manager, second)
	if got := Gia_ManReadSeqModelVec(manager); got != second {
		t.Fatal("set accessor did not transfer the supplied vector")
	}
	if got := Gia_ManTakeSeqModelVec(manager); got != second {
		t.Fatal("take accessor did not return the manager's vector")
	}
	if got := Gia_ManReadSeqModelVec(manager); got != nil {
		t.Fatal("take accessor did not clear the manager's vector")
	}
	if got := Gia_ManTakeSeqModelVec(manager); got != nil {
		t.Fatal("take accessor from an empty manager should return nil")
	}

}

func TestSeqModelVecAccessorsAreNilSafe(t *testing.T) {
	var manager *Gia_Man_t
	if Gia_ManReadSeqModelVec(manager) != nil || Gia_ManTakeSeqModelVec(manager) != nil {
		t.Fatal("read and take on nil manager should return nil")
	}
	Gia_ManSetSeqModelVec(manager, nil)
}
