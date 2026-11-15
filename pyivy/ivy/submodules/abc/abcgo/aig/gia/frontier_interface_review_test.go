package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestFrontierInterfaceOwnership(t *testing.T) {
	g := frontierReviewGraph(7)
	defer Gia_ManStop(g)
	g.pName, g.PSpec = "frontier", "frontier.aig"
	// Change logical order without moving physical objects or register roles.
	for _, ids := range []*Vec_Int_t{g.vCis, g.vCos} {
		a, b := Vec_IntEntry(ids, 0), Vec_IntEntry(ids, 1)
		Vec_IntWriteEntry(ids, 0, b)
		Vec_IntWriteEntry(ids, 1, a)
		Gia_ObjSetCioId(g.pObjs[b], 0)
		Gia_ObjSetCioId(g.pObjs[a], 1)
	}
	if err := g.SetAigerProperties(1, 1); err != nil {
		t.Fatal(err)
	}
	g.aigerLatchInits = []int32{1}
	g.aigerComment = []byte("frontier interface")
	g.VNamesIn, g.VNamesOut = Vec_PtrAlloc(5), Vec_PtrAlloc(3)
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		Vec_PtrPush(g.VNamesIn, fmt.Sprintf("ci%d", i))
	}
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		Vec_PtrPush(g.VNamesOut, fmt.Sprintf("co%d", i))
	}
	out := Gia_ManFront(g)
	defer Gia_ManStop(out)
	Gia_ManFrontTransform(out)
	cofReviewCheckRebuild(t, g, out)
	if out.pName != g.pName || out.PSpec != g.PSpec || out.nConstrs != g.nConstrs || out.aigerBadOutputs != g.aigerBadOutputs || !slices.Equal(out.aigerLatchInits, g.aigerLatchInits) || !slices.Equal(out.aigerComment, g.aigerComment) || !slices.Equal(out.InputNames(), g.InputNames()) || !slices.Equal(out.OutputNames(), g.OutputNames()) {
		t.Fatal("frontier conversion lost interface metadata")
	}
	if !slices.Equal(Vec_IntArray(out.vCis), Vec_IntArray(g.vCis)) || !slices.Equal(Vec_IntArray(out.vCos), Vec_IntArray(g.vCos)) {
		t.Fatal("logical ports no longer reference original object IDs")
	}
	out.aigerLatchInits[0] = 0
	out.aigerComment[0] = 'X'
	Vec_PtrWriteEntry(out.VNamesIn, 0, "changed")
	if g.aigerLatchInits[0] != 1 || string(g.aigerComment) != "frontier interface" || g.InputNames()[0] != "ci0" {
		t.Fatal("frontier interface aliases source")
	}
}
