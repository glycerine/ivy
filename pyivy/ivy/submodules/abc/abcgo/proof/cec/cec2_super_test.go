package cec

import (
	"reflect"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestCec2SuperStopsAtSharedAndComplementedNodes(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	shared := gia.Gia_ManAppendAnd(g, x, y)
	root := gia.Gia_ManAppendAnd(g, shared, z^1)
	gia.Gia_ManAppendCo(g, root)
	gia.Gia_ManAppendCo(g, shared)
	gia.Gia_ManCreateRefs(g)
	check := func(got, want []int32) bool { slices.Sort(got); slices.Sort(want); return reflect.DeepEqual(got, want) }
	if got := cec2CollectSuper(g, root, false); !check(got, []int32{shared, z ^ 1}) {
		t.Fatalf("shared boundary: %v", got)
	}
	g.Refs[shared>>1] = 1
	if got := cec2CollectSuper(g, root, false); !check(got, []int32{x, y, z ^ 1}) {
		t.Fatalf("flattened boundary: %v", got)
	}
	gia.Gia_ObjSetMark0(gia.Gia_ManObj(g, shared>>1), 1)
	if got := cec2CollectSuper(g, root, true); !check(got, []int32{shared, z ^ 1}) {
		t.Fatalf("marked mux boundary: %v", got)
	}
}
