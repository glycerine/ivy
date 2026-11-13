package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestArtificialBoxesPreserveNativeBuffer(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	buffer := gia.Gia_ManAppendBuf(g, x^1)
	gia.Gia_ManAppendCo(g, buffer)
	boxed := Gia_ManDupWithArtificialBoxes(g, 10, 1, 8, 1, 0, 0, 1, 0)
	timing, extra := boxed.BoxTiming()
	if tim.Tim_ManBoxNum(timing) != 0 {
		t.Fatal("buffer was annotated as a carry AND")
	}
	collapsed := gia.Gia_ManDupCollapse(boxed, extra, nil, 0)
	for _, v := range []bool{false, true} {
		got, err := collapsed.Evaluate([]bool{v})
		if err != nil || len(got) != 1 || got[0] == v {
			t.Fatalf("input=%v got=%v err=%v", v, got, err)
		}
	}
}
