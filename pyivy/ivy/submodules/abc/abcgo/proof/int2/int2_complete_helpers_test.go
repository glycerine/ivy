package int2

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// The C cube loop initializes each cube to true, even for an empty cube.
func TestSourceEmptyCubeMakesSOPTrue(t *testing.T) {
	g := gia.Gia_ManStart(8)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManAppendCo(g, state^1)
	gia.Gia_ManSetRegNum(g, 1)
	sop := vec.Vec_IntAllocArrayCopy([]int32{0, 0, 1, 0}, 4)
	out := Int2_ManProbToGia(g, sop)
	defer gia.Gia_ManStop(out)
	for _, bit := range []bool{false, true} {
		result, err := out.Evaluate([]bool{bit})
		if err != nil || len(result) != 1 || !result[0] {
			t.Fatalf("state %v: result %v, error %v; want true", bit, result, err)
		}
	}
}

// Both complete C helpers assert that at least one register is present.
func TestSourceInitHelpersRejectCombinationalGraph(t *testing.T) {
	for _, run := range []struct {
		name string
		f    func(*Gia_Man_t)
	}{
		{"dup", func(g *Gia_Man_t) { Int2_ManDupInit(g, 0) }},
		{"check", func(g *Gia_Man_t) { Int2_ManCheckInit(g) }},
	} {
		t.Run(run.name, func(t *testing.T) {
			g := gia.Gia_ManStart(4)
			gia.Gia_ManAppendCo(g, 0)
			defer gia.Gia_ManStop(g)
			defer func() {
				if recover() == nil {
					t.Fatal("expected source assertion")
				}
			}()
			run.f(g)
		})
	}
}
