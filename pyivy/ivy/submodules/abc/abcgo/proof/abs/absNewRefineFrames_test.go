package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestNewRefineMultiFrameRealCounterexample(t *testing.T) {
	// q(0)=0, q(t+1)=x(t), bad(t)=x(t)&q(t). The trace x=1
	// reaches q=1 at frame 1 and violates bad in frame 2. Starting
	// refinement at frame 1 exercises the source's state transformation.
	g := gia.Gia_ManStart(12)
	x := gia.Gia_ManAppendCi(g)
	q := gia.Gia_ManAppendCi(g)
	bad := gia.Gia_ManAppendAnd(g, x, q)
	gia.Gia_ManAppendCo(g, bad)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 1)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	for _, id := range []int32{0, q >> 1, bad >> 1} {
		vec.Vec_IntWriteEntry(classes, int(id), 1)
	}
	gia.Gia_ManSetGateClasses(g, classes)
	cex := util.Abc_CexAlloc(1, 1, 3)
	cex.IFrame = 2
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		util.Abc_InfoSetBit(cex.PData, 1+frame)
	}
	if gia.Gia_ManVerifyCex(g, cex, 0) != 1 {
		t.Fatal("fixture is not a real counterexample")
	}
	state := Gia_ManGetStateAndCheckCex(g, cex, 1)
	if state == nil || vec.Vec_IntSize(state) != 1 || vec.Vec_IntEntry(state, 0) != 1 {
		t.Fatal("frame 1 must start with q=1")
	}
	flops := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(flops, q>>1)
	shifted := Gia_ManTransformFlops(g, flops, state)
	_, holds := replayCexFromZero(shifted, cex, 1, 1)
	if !holds {
		t.Fatal("fast-forwarded model must preserve the suffix counterexample")
	}
	gia.Gia_ManSetGateClasses(shifted, vec.Vec_IntDup(classes))
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	pars.nFramesMax = 2 // iFrame - frameStart + 1, as in absOut.c.
	if result := Gia_ManPerformGla(shifted, &pars); result != 0 {
		t.Fatalf("fast-forwarded real counterexample result = %d, want 0", result)
	}
	if got := Gia_ManNewRefine(g, cex, 1, 0, 0); got != -1 {
		t.Fatalf("Gia_ManNewRefine returned %d, want -1", got)
	}
	got := gia.Gia_ManGateClasses(g)
	if got == nil || vec.Vec_IntSize(got) != vec.Vec_IntSize(classes) {
		t.Fatal("refinement lost abstraction classes")
	}
	for i := 0; i < vec.Vec_IntSize(classes); i++ {
		if vec.Vec_IntEntry(got, i) != vec.Vec_IntEntry(classes, i) {
			t.Fatalf("class %d = %d, want %d", i, vec.Vec_IntEntry(got, i), vec.Vec_IntEntry(classes, i))
		}
	}
}
