package wln

import (
	"reflect"
	"testing"
)

func TestFaninVectorSourceStorageTransitions(t *testing.T) {
	var v Wln_Vec_t
	for i := int32(0); i < 10; i++ {
		v.appendFanin(10 + i)
		want := make([]int32, i+1)
		for j := range want {
			want[j] = 10 + int32(j)
		}
		if !reflect.DeepEqual(v.fanins(), want) {
			t.Fatalf("after %d: %v want %v", i+1, v.fanins(), want)
		}
		wantCap := int32(0)
		if i >= 2 {
			wantCap = 4
		}
		if i >= 4 {
			wantCap = 8
		}
		if i >= 8 {
			wantCap = 16
		}
		if v.nCap != wantCap {
			t.Fatalf("after %d: cap %d want %d", i+1, v.nCap, wantCap)
		}
	}
	v.fanins()[1] = 77
	if v.fanins()[1] != 77 {
		t.Fatal("fanin slice must allow in-place updates")
	}
}
