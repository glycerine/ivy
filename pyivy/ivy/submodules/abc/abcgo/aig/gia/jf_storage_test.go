package gia

import (
	"reflect"
	"testing"
)

func TestJfCutMemoryAndBestCut(t *testing.T) {
	m := jfCutMemory{}
	initial := []int32{3, 2, 2, 4, 1, 6, 3, 8, 10, 12}
	id := m.append(initial, len(initial))
	list := m.get(id)
	initial[1] = 0
	for i := 0; i < 200; i++ {
		m.append([]int32{1, 1, int32(i)}, 3)
	}
	if list[1] != 2 {
		t.Fatal("memory did not copy input")
	}
	selected := list[4:]
	Jf_ObjSetBestCut(list, selected, nil)
	want := []int32{3, 1, 6, 2, 2, 4, 3, 8, 10, 12}
	if !reflect.DeepEqual(m.get(id), want) {
		t.Fatalf("cut order %v", m.get(id))
	}
	Jf_ObjSetBestCut(list, list[1:], nil)
	if !reflect.DeepEqual(list, want) {
		t.Fatal("front cut changed")
	}
	m.reset()
	if len(m.lists) != 0 || m.memory() != 0 {
		t.Fatal("reset failed")
	}
}
