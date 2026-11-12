package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestKfCutStorageRoundTrip(t *testing.T) {
	a, b := &Kf_Cut_t{nLeaves: 2, iFunc: 4, Delay: 3, Area: 1.5}, &Kf_Cut_t{nLeaves: 1, iFunc: 2, Delay: 1, Area: 2.5}
	a.pLeaves[0], a.pLeaves[1], b.pLeaves[0] = 1, 4, 7
	v := vec.Vec_IntAlloc(0)
	Kf_ManSaveResults([]*Kf_Cut_t{a, b}, 2, b, v)
	if vec.Vec_IntEntry(v, 1) != 8 {
		t.Fatalf("best offset: %v", vec.Vec_IntArray(v))
	}
	var out [3]Kf_Cut_t
	if Kf_SetLoadCuts(out[:], vec.Vec_IntArray(v)) != 2 || out[0].Area != 1.5 || out[1].Delay != 1 || !reflect.DeepEqual(out[0].pLeaves[:2], []int32{1, 4}) {
		t.Fatalf("round trip %+v", out)
	}
	Kf_ManStoreAddUnit(v, 9, 5, 3.5)
	if Kf_SetLoadCuts(out[:], vec.Vec_IntArray(v)) != 3 || out[2].pLeaves[0] != 9 {
		t.Fatal("unit cut")
	}
}
func TestKfCutMerging(t *testing.T) {
	a, b := &Kf_Cut_t{nLeaves: 2}, &Kf_Cut_t{nLeaves: 2}
	a.pLeaves[0], a.pLeaves[1] = 1, 4
	b.pLeaves[0], b.pLeaves[1] = 4, 8
	for _, merge := range []func(*Kf_Cut_t, *Kf_Cut_t, *Kf_Cut_t, int32) int32{Kf_SetMergeSimpleOne, Kf_SetMergeOrderOne} {
		var out Kf_Cut_t
		if merge(a, b, &out, 3) != 1 || !reflect.DeepEqual(out.pLeaves[:out.nLeaves], []int32{1, 4, 8}) {
			t.Fatal("merge", out)
		}
		if merge(a, b, &out, 2) != 0 {
			t.Fatal("limit")
		}
		if !Kf_SetCutIsContainedSimple(&out, a) || !Kf_SetCutIsContainedOrder(&out, a) {
			t.Fatal("containment")
		}
	}
}
