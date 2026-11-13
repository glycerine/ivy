package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"testing"
)

func TestImpStateSpaceRatioCountsUnionOfViolations(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	p := &Fra_Man_t{PManAig: g, PCla: &classPartition{}}
	if got := Fra_ImpComputeStateSpaceRatio(p); got != 0 {
		t.Fatalf("no implications = %g", got)
	}
	p.PCla.VImps = vec.Vec_IntAlloc(3)
	vec.Vec_IntPush(p.PCla.VImps, Fra_ImpCreate(int32(x.NodeID()), 0))
	vec.Vec_IntPush(p.PCla.VImps, Fra_ImpCreate(int32(x.NodeID()), 0))
	vec.Vec_IntPush(p.PCla.VImps, Fra_ImpCreate(int32(x.NodeID()), int32(x.NodeID())))
	seed := aig.NewAigRandomSource()
	seed.Uint32()
	count := 0
	for i := 0; i < 64; i++ {
		count += bits.OnesCount32(seed.Uint32())
	}
	aig.Aig_ManRandom(1)
	want := 100 * float64(count) / (64 * 32)
	if got := Fra_ImpComputeStateSpaceRatio(p); got != want {
		t.Fatalf("violated ratio = %g, want %g", got, want)
	}
}
