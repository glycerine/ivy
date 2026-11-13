package ssc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
)

// Ssc_ManStart calls Ssc_GiaClassesInit before returning; sscClass.c sets
// each CI/AND candidate's representative to the constant-zero object.
func TestSourceManagerStartPublishesCandidateMetadata(t *testing.T) {
	subject := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(subject), gia.Gia_ManAppendCi(subject)
	node := gia.Gia_ManAppendAnd(subject, a, b)
	output := gia.Gia_ManAppendCo(subject, node)
	care := gia.Gia_ManStart(10)
	gia.Gia_ManAppendCi(care)
	gia.Gia_ManAppendCi(care)
	gia.Gia_ManAppendCo(care, 0)
	manager, status := Ssc_ManStart(subject, care, nil)
	if status != satoko.SATOKO_SAT || manager == nil {
		t.Fatalf("status %d manager %v", status, manager)
	}
	if !subject.HasRepresentatives() {
		t.Fatal("source candidates were not initialized on manager start")
	}
	for _, id := range []int32{a >> 1, b >> 1, node >> 1} {
		if got := gia.Gia_ObjRepr(subject, id); got != 0 {
			t.Errorf("candidate %d representative = %d, want constant 0", id, got)
		}
	}
	if got := gia.Gia_ObjRepr(subject, output>>1); got != gia.GIA_VOID {
		t.Errorf("output representative = %d, want void", got)
	}
}

// Ssc_GiaClassesRefine updates the GIA's representative and linked-list
// metadata immediately, not only after the outer sweep finishes.
func TestSourceClassRefinementPublishesMetadata(t *testing.T) {
	graph := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(graph), gia.Gia_ManAppendCi(graph)
	gia.Gia_ManAppendAnd(graph, a, b)
	classes := Ssc_GiaClassesInit(graph, []bool{false, false})
	sim := NewSimulation(graph, 1)
	Ssc_GiaSavePiPattern(sim, []bool{false, false})
	Ssc_GiaSavePiPattern(sim, []bool{true, false})
	Ssc_GiaSimRound(sim)
	Ssc_GiaClassesRefine(classes, sim)
	if !graph.HasRepresentatives() {
		t.Fatal("class metadata is absent after refinement")
	}
	for id, repr := range classes.Repr {
		want := repr
		if want < 0 {
			want = gia.GIA_VOID
		}
		if got := gia.Gia_ObjRepr(graph, int32(id)); got != want {
			t.Errorf("object %d representative = %d, want %d", id, got, want)
		}
	}
	for _, group := range classes.Groups {
		for i, id := range group {
			want := int32(0)
			if group[0] != 0 && i+1 < len(group) {
				want = group[i+1]
			}
			if got := gia.Gia_ObjNext(graph, id); got != want {
				t.Errorf("object %d next = %d, want %d", id, got, want)
			}
		}
	}
}

func TestPublishClassMetadataKeepsProvedFlags(t *testing.T) {
	graph := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(graph), gia.Gia_ManAppendCi(graph)
	graph.StartRepresentatives()
	gia.Gia_ObjSetProved(graph, b>>1)
	classes := &Classes{owner: graph, Repr: make([]int32, graph.ObjectCount()), Groups: [][]int32{{0}, {a >> 1, b >> 1}}}
	classes.publish()
	if got := gia.Gia_ObjRepr(graph, b>>1); got != a>>1 {
		t.Fatalf("representative = %d", got)
	}
	if got := gia.Gia_ObjNext(graph, a>>1); got != b>>1 {
		t.Fatalf("head next = %d", got)
	}
	if got := gia.Gia_ObjProved(graph, b>>1); got != 1 {
		t.Fatal("proved flag was cleared")
	}
	classes.Groups = [][]int32{{0, a >> 1}, {b >> 1}}
	classes.publish()
	if got := gia.Gia_ObjRepr(graph, a>>1); got != 0 {
		t.Fatalf("constant candidate representative = %d", got)
	}
	if got := gia.Gia_ObjNext(graph, a>>1); got != 0 {
		t.Fatalf("obsolete class link = %d", got)
	}
	if got := gia.Gia_ObjRepr(graph, b>>1); got != gia.GIA_VOID {
		t.Fatalf("singleton representative = %d", got)
	}
	if got := gia.Gia_ObjProved(graph, b>>1); got != 1 {
		t.Fatal("proved flag was cleared during refinement")
	}
}
