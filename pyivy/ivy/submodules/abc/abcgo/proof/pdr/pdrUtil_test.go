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
	if printed.String() != "0-1-0" { t.Fatalf("pattern %q", printed.String()) }
	out := vec.Vec_StrAlloc(0)
	Pdr_SetPrintStr(out, a, 5, nil)
	if string(vec.Vec_StrArray(out)) != "0-1-0 1\n" { t.Fatalf("cube output %q", vec.Vec_StrArray(out)) }
	keep := hash.Hash_IntAlloc(7)
	if got := ZPdr_SetIntersection(a, b, keep); got == nil || got.nLits != 2 { t.Fatalf("intersection %+v", got) }
	hash.Hash_IntWriteEntry(keep, 9, 1)
	if ZPdr_SetIntersection(a, b, keep) != nil { t.Fatal("lost required trailing literal") }
}

func TestPdrObligationQueueOrderAndReferences(t *testing.T) {
	manager := &Pdr_Man_t{}
	items := []*Pdr_Obl_t{
		Pdr_OblStart(2, 0, testPdrSet([]int32{1}, nil), nil),
		Pdr_OblStart(1, 3, testPdrSet([]int32{3}, nil), nil),
		Pdr_OblStart(1, 2, testPdrSet([]int32{5}, nil), nil),
	}
	for _, item := range items { Pdr_QueuePush(manager, item) }
	for _, want := range []*Pdr_Obl_t{items[2], items[1], items[0]} {
		got := Pdr_QueuePop(manager)
		if got != want || got.nRefs != 1 { t.Fatalf("got %+v, want %+v", got, want) }
		Pdr_OblDeref(got)
	}
	if manager.nQueCur != 0 || manager.nQueMax != 3 || Pdr_QueueIsEmpty(manager) != 1 { t.Fatalf("queue state %+v", manager) }
}
