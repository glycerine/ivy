package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// absRpm.c expands a shared support node when either of its fanins is
// already in the dereferenced cone; it removes the node if both are present.
func TestExtendedSupportExpandsSharedNodes(t *testing.T) {
	for _, bothSeen := range []bool{false, true} {
		p := gia.Gia_ManStart(16)
		a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		d := gia.Gia_ManAppendCi(p)
		shared := gia.Gia_ManAppendAnd(p, a, b)
		var other int32
		if bothSeen {
			other = gia.Gia_ManAppendAnd(p, gia.Gia_ManAppendAnd(p, a, b), d)
		} else {
			other = gia.Gia_ManAppendAnd(p, a, c)
		}
		root := gia.Gia_ManAppendAnd(p, shared, other)
		gia.Gia_ManAppendCo(p, root)
		gia.Gia_ManAppendCo(p, shared)
		for _, lit := range []int32{a, b, c, d} {
			gia.Gia_ObjSetMark1(gia.Gia_ManObj(p, lit>>1), 1)
		}
		gia.Gia_ManCreateRefs(p)
		r := gia.Gia_ManObj(p, root>>1)
		deref := Abs_GiaObjDeref_rec(p, r)
		support := vec.Vec_IntAlloc(4)
		count := Abs_ManSupport3(p, r, support)
		seen := make(map[int32]bool)
		for _, id := range vec.Vec_IntArray(support) {
			seen[id] = true
		}
		want := 3
		if int(count) != want || len(seen) != want || !seen[a>>1] || !seen[b>>1] || seen[shared>>1] || (!bothSeen && !seen[c>>1]) || (bothSeen && !seen[d>>1]) {
			t.Fatalf("bothSeen=%v support=%v", bothSeen, vec.Vec_IntArray(support))
		}
		if restored := Abs_GiaObjRef_rec(p, r); restored != deref {
			t.Fatalf("bothSeen=%v dereferenced %d nodes, restored %d", bothSeen, deref, restored)
		}
	}
}
