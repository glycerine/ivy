package pdr

import (
	"bytes"
	"testing"

	"github.com/glycerine/abcgo/misc/hash"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestPdrSetPrintingAndIntersection(t *testing.T) {
	a, b := testPdrSet([]int32{1, 4, 9}, nil), testPdrSet([]int32{1, 4}, nil)
	var printed bytes.Buffer
	Pdr_SetPrint(&printed, a, 5, nil)
	if printed.String() != "0-1-0" {
		t.Fatalf("pattern %q", printed.String())
	}
	out := vec.Vec_StrAlloc(0)
	Pdr_SetPrintStr(out, a, 5, nil)
	if string(vec.Vec_StrArray(out)) != " 1\n" {
		t.Fatalf("cube output %q", vec.Vec_StrArray(out))
	}
	filtered := vec.Vec_StrAlloc(0)
	Pdr_SetPrintStr(filtered, a, 5, vec.Vec_IntAllocArrayCopy([]int32{1, 0, 1, 0, 1}, 5))
	if string(vec.Vec_StrArray(filtered)) != "010 1\n" {
		t.Fatalf("filtered cube output %q", vec.Vec_StrArray(filtered))
	}
	keep := hash.Hash_IntAlloc(7)
	if got := ZPdr_SetIntersection(a, b, keep); got == nil || got.nLits != 2 {
		t.Fatalf("intersection %+v", got)
	}
	hash.Hash_IntWriteEntry(keep, 9, 1)
	if got := ZPdr_SetIntersection(a, b, keep); got == nil || got.nLits != 2 {
		t.Fatal("C ignores unmatched trailing literals")
	}
	hash.Hash_IntWriteEntry(keep, 3, 1)
	if ZPdr_SetIntersection(testPdrSet([]int32{1, 3, 4}, nil), b, keep) != nil {
		t.Fatal("matched-range required literal was dropped")
	}
}

func TestPdrObligationQueueOrderAndReferences(t *testing.T) {
	manager := &Pdr_Man_t{}
	items := []*Pdr_Obl_t{
		Pdr_OblStart(2, 0, testPdrSet([]int32{1}, nil), nil),
		Pdr_OblStart(1, 3, testPdrSet([]int32{3}, nil), nil),
		Pdr_OblStart(1, 2, testPdrSet([]int32{5}, nil), nil),
	}
	for _, item := range items {
		Pdr_QueuePush(manager, item)
	}
	for _, want := range []*Pdr_Obl_t{items[2], items[1], items[0]} {
		got := Pdr_QueuePop(manager)
		if got != want || got.nRefs != 1 {
			t.Fatalf("got %+v, want %+v", got, want)
		}
		Pdr_OblDeref(got)
	}
	if manager.nQueCur != 0 || manager.nQueMax != 3 || Pdr_QueueIsEmpty(manager) != 1 {
		t.Fatalf("queue state %+v", manager)
	}
}

// C inserts equal-priority obligations after existing peers, and a queue
// cleanup releases each obligation's reference to its shared successor.
func TestPdrQueueStableTiesAndSharedSuccessor(t *testing.T) {
	manager := &Pdr_Man_t{}
	next := Pdr_OblStart(2, 0, testPdrSet([]int32{1}, nil), nil)
	first := Pdr_OblStart(1, 4, testPdrSet([]int32{3}, nil), next)
	second := Pdr_OblStart(1, 4, testPdrSet([]int32{5}, nil), Pdr_OblRef(next))
	before := Pdr_OblStart(1, 3, testPdrSet([]int32{7}, nil), nil)
	for _, item := range []*Pdr_Obl_t{first, second, before} {
		Pdr_QueuePush(manager, item)
	}
	if got := Pdr_QueuePop(manager); got != before {
		t.Fatalf("first popped %p, want %p", got, before)
	} else {
		Pdr_OblDeref(got)
	}
	if got := Pdr_QueuePop(manager); got != first {
		t.Fatalf("equal-priority queue order got %p, want %p", got, first)
	} else {
		Pdr_OblDeref(got)
	}
	if next.nRefs != 1 || next.pState == nil {
		t.Fatalf("shared successor released early: %+v", next)
	}
	Pdr_QueueStop(manager)
	if manager.nQueCur != 0 || manager.pQueue != nil || second.nRefs != 0 || next.nRefs != 0 || next.pState != nil {
		t.Fatalf("queue stop retained obligations: manager=%+v second=%+v next=%+v", manager, second, next)
	}
}
