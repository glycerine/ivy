package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGiaIffClustersPairAndTracksTiming(t *testing.T) {
	graph := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	x := gia.Gia_ManAppendAnd(graph, a, b)
	y := gia.Gia_ManAppendAnd(graph, x, c)
	Gia_ManAppendCo(graph, y)
	mapping := vec.Vec_IntStart(int(Gia_ManObjNum(graph)))
	for _, record := range []struct {
		id     int32
		leaves []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}} {
		vec.Vec_IntWriteEntry(mapping, int(record.id), int32(Vec_IntSize(mapping)))
		Vec_IntPush(mapping, int32(len(record.leaves)))
		for _, leaf := range record.leaves {
			Vec_IntPush(mapping, leaf)
		}
		Vec_IntPush(mapping, record.id)
	}
	if err := graph.SetMappedData(mapping, nil, nil, ""); err != nil {
		t.Fatal(err)
	}
	lib := &If_LibLut_t{LutMax: 7}
	for i := range lib.pLutDelays {
		lib.pLutDelays[i][0] = 1
	}
	timing := tim.Tim_ManStart(3, 1)
	defer tim.Tim_ManStop(timing)
	for i := 0; i < 3; i++ {
		tim.Tim_ManInitPiArrival(timing, i, 0)
	}
	p := Gia_ManIffPerform(graph, lib, timing, 4, 2)
	defer Gia_ManIffStop(p)
	if got := Iff_ObjMatchId(p, y>>1, 2); got != x>>1 {
		t.Fatalf("pair match %d, want %d", got, x>>1)
	}
	if got := Iff_ObjTimeId(p, y>>1); got != 1 {
		t.Fatalf("root arrival %v, want 1", got)
	}
	packing := Gia_ManIffSelect(p)
	want := []int32{1, 2, x >> 1, y >> 1}
	got := Vec_IntArray(packing)
	if len(got) != len(want) {
		t.Fatalf("packing %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("packing %v want %v", got, want)
		}
	}
	if err := graph.ReplacePacking(packing); err != nil {
		t.Fatal(err)
	}
}

func TestGiaIffClustersTriple(t *testing.T) {
	graph := Gia_ManStart(16)
	a, b, c, d := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph), Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	x, y := gia.Gia_ManAppendAnd(graph, a, b), gia.Gia_ManAppendAnd(graph, c, d)
	root := gia.Gia_ManAppendAnd(graph, x, y)
	Gia_ManAppendCo(graph, root)
	mapping := vec.Vec_IntStart(int(Gia_ManObjNum(graph)))
	for _, rec := range []struct {
		id     int32
		leaves []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{c >> 1, d >> 1}}, {root >> 1, []int32{x >> 1, y >> 1}}} {
		vec.Vec_IntWriteEntry(mapping, int(rec.id), int32(Vec_IntSize(mapping)))
		Vec_IntPush(mapping, 2)
		for _, v := range rec.leaves {
			Vec_IntPush(mapping, v)
		}
		Vec_IntPush(mapping, rec.id)
	}
	if err := graph.SetMappedData(mapping, nil, nil, ""); err != nil {
		t.Fatal(err)
	}
	lib := &If_LibLut_t{LutMax: 10}
	for i := range lib.pLutDelays {
		lib.pLutDelays[i][0] = 1
	}
	timing := tim.Tim_ManStart(4, 1)
	defer tim.Tim_ManStop(timing)
	p := Gia_ManIffPerform(graph, lib, timing, 4, 3)
	defer Gia_ManIffStop(p)
	if p.matches[2][root>>1] != x>>1 || p.matches[3][root>>1] != y>>1 {
		t.Fatalf("matches %d %d", p.matches[2][root>>1], p.matches[3][root>>1])
	}
	got := Vec_IntArray(Gia_ManIffSelect(p))
	want := []int32{1, 3, x >> 1, y >> 1, root >> 1}
	if len(got) != len(want) {
		t.Fatalf("packing %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("packing %v want %v", got, want)
		}
	}
}
