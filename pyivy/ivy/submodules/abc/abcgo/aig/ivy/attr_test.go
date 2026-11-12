package ivy

import (
	"reflect"
	"testing"
)

func TestAttrIntegersGrowAndKeepSpots(t *testing.T) {
	p := Attr_ManStartInt(1)
	slot := Attr_ManFetchSpotInt(p, 0)
	*slot = 42
	far := Attr_ManFetchSpotInt(p, 10000)
	if *far != 0 || Attr_ManReadAttrInt(p, 0) != 42 {
		t.Fatal("growth lost data")
	}
	*slot = 99
	if Attr_ManReadAttrInt(p, 0) != 99 {
		t.Fatal("slot address invalidated")
	}
	for _, id := range []int32{1, 50, 9999} {
		if Attr_ManReadAttrInt(p, id) != 0 {
			t.Fatal("new slot not zero")
		}
	}
	Attr_ManWriteAttrInt(p, 2, -7)
	if Attr_ManFetchAttrInt(p, 2) != -7 {
		t.Fatal("integer write")
	}
	Attr_ManStop(p)
}
func TestAttrPointersAndManagedMemory(t *testing.T) {
	p := Attr_ManStartPtr(0)
	value := new(int)
	*value = 7
	slot := Attr_ManFetchSpotPtr(p, 4)
	*slot = value
	Attr_ManSetAttrPtr(p, 5000, "far")
	if Attr_ManReadAttrPtr(p, 4) != value {
		t.Fatal("pointer lost")
	}
	*slot = "updated"
	if Attr_ManReadAttrPtr(p, 4) != "updated" {
		t.Fatal("pointer slot moved")
	}
	var nilPointer *int
	Attr_ManSetAttrPtr(p, 9, nilPointer)
	if Attr_ManReadAttrPtr(p, 9) != nil {
		t.Fatal("typed nil not normalized")
	}
	managed := Attr_ManStartPtrMem(2, 8)
	first := Attr_ManFetchAttrPtr(managed, 0).([]byte)
	first[0] = 99
	if again := Attr_ManFetchAttrPtr(managed, 0).([]byte); again[0] != 99 || &again[0] != &first[0] {
		t.Fatal("managed fetch reallocated existing object")
	}
	far := Attr_ManFetchAttrPtr(managed, 4000).([]byte)
	if len(far) != 8 || !reflect.DeepEqual(far, make([]byte, 8)) {
		t.Fatal("managed growth not initialized")
	}
	Attr_ManStop(p)
	Attr_ManStop(managed)
}
func TestAttrCleanupOrderAndOnce(t *testing.T) {
	p := Attr_ManStartPtr(3)
	Attr_ManWriteAttrPtr(p, 0, "a")
	Attr_ManWriteAttrPtr(p, 2, "b")
	var calls []string
	p.SetCleanup("ctx", func(ctx any) { calls = append(calls, ctx.(string)) }, func(ctx, value any) { calls = append(calls, value.(string)); Attr_ManStop(p) })
	Attr_ManStop(p)
	Attr_ManStop(p)
	if !reflect.DeepEqual(calls, []string{"a", "b", "ctx"}) {
		t.Fatal(calls)
	}
	ints := Attr_ManStartInt(2)
	Attr_ManWriteAttrInt(ints, 1, -3)
	var got int32
	ints.SetCleanup(nil, nil, func(_ any, value any) { got = value.(int32) })
	Attr_ManStop(ints)
	if got != -3 {
		t.Fatal(got)
	}
}
func TestAttrInvalidOperations(t *testing.T) {
	cases := []func(){func() { Attr_ManFetchAttrInt(Attr_ManStartInt(0), -1) }, func() { Attr_ManReadAttrInt(Attr_ManStartPtr(0), 0) }, func() { Attr_ManWriteAttrPtr(Attr_ManStartPtrMem(1, 2), 0, "replace") }, func() { Attr_ManWriteAttrInt(Attr_ManStartInt(0), 0, 1) }, func() { p := Attr_ManStartInt(1); Attr_ManStop(p); Attr_ManReadAttrInt(p, 0) }}
	for i, f := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("case %d did not panic", i)
				}
			}()
			f()
		}()
	}
}
