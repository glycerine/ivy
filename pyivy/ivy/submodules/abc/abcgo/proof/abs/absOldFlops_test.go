package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestFlopClassesRoundTrip(t *testing.T) {
	g := gia.Gia_ManStart(16)
	for i := 0; i < 4; i++ {
		gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, 0)
	}
	gia.Gia_ManSetRegNum(g, 4)
	classes := vec.Vec_IntAllocArrayCopy([]int32{0, 3, 0, -1}, 4)
	flops := Gia_ManClasses2Flops(classes)
	if got := vec.Vec_IntArray(flops); !reflect.DeepEqual(got, []int32{1, 3}) {
		t.Fatalf("flop IDs %v", got)
	}
	if got := vec.Vec_IntArray(Gia_ManFlops2Classes(g, flops)); !reflect.DeepEqual(got, []int32{0, 1, 0, 1}) {
		t.Fatalf("class map %v", got)
	}
}
