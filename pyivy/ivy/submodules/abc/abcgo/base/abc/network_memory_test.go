package abc

import (
	"testing"
	"unsafe"
)

func TestNtkMemoryVectorBytesUsesCapacities(t *testing.T) {
	n := NewNetwork("memory-capacity")
	n.kindOrder[ObjectPI] = make([]*Object, 1, 11)
	n.kindOrder[ObjectPO] = make([]*Object, 2, 13)
	n.cis = make([]*Object, 1, 17)
	n.cos = make([]*Object, 1, 19)
	n.objects = make([]*Object, 1, 23)
	obj := &Object{network: n, fanins: make([]Edge, 1, 7), fanouts: make([]*Object, 1, 5)}
	n.objects[0] = obj

	want := uintptr(11+13+17+19+23)*unsafe.Sizeof((*Object)(nil)) +
		uintptr(7+5)*unsafe.Sizeof(int32(0))
	if got := Abc_NtkMemoryVectorBytes(n); got != float64(want) {
		t.Fatalf("memory bytes = %v, want capacity-based %d", got, want)
	}
	if got := Abc_NtkMemoryVectorBytes(nil); got != 0 {
		t.Fatalf("nil network memory = %v, want 0", got)
	}
}

func TestNtkBarBufs2Accessor(t *testing.T) {
	n := NewNetwork("secondary-barbuf-count")
	if Abc_NtkBarBufs2(n) != 0 {
		t.Fatal("new network has a nonzero secondary barrier-buffer count")
	}
	Abc_NtkIncrementBarBufs2(n)
	Abc_NtkIncrementBarBufs2(n)
	if got := Abc_NtkBarBufs2(n); got != 2 {
		t.Fatalf("secondary barrier-buffer count = %d, want 2", got)
	}
}
