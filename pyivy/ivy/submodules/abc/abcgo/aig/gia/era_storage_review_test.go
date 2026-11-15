package gia

import "testing"

func TestExplicitReachabilityDataViewBoundary(t *testing.T) {
	g := eraReviewGraph(6, 1, 0)
	defer Gia_ManStop(g)
	p := Gia_ManEraCreate(g)
	defer Gia_ManEraFree(p)
	id := Gia_ManPi(g, 0).id
	next := Gia_ManEraData(p, id+1)
	before := append([]uint32(nil), next...)
	row := Gia_ManEraData(p, id)
	// A borrowed object row may be edited in place, but append must not
	// overwrite the next object's simulation words.
	row[0] = 123
	if Gia_ManEraData(p, id)[0] != 123 {
		t.Fatal("row no longer aliases object")
	}
	appended := append(row, 0xdeadbeef)
	if appended[len(row)] != 0xdeadbeef {
		t.Fatal("append failed")
	}
	for i, v := range next {
		if v != before[i] {
			t.Fatalf("append overwrote neighboring object word %d: got %x, want %x", i, v, before[i])
		}
	}
}
