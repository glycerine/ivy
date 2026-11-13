package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// fraBmc.c checks implication candidates from nPref through nFramesAll.
func TestBMCImplicationsSkipPrefixFrames(t *testing.T) {
	original := aig.NewGraph()
	x, y := original.NewInput(), original.NewInput()
	frames := aig.NewGraph()
	fx, fy := frames.NewInput(), frames.NewInput()
	reduced := aig.NewGraph()
	rep := reduced.NewInput()
	b := Fra_BmcStart(original, 1, 2)
	b.PAigFrames = frames
	b.VImps = vec.Vec_IntAlloc(1)
	Bmc_ObjSetFrames(b, x, 0, fx)
	Bmc_ObjSetFrames(b, y, 0, fy)
	for frame := int32(1); frame < 3; frame++ {
		Bmc_ObjSetFrames(b, x, frame, fx)
		Bmc_ObjSetFrames(b, y, frame, fx)
	}
	p := &Fra_Man_t{PManAig: frames, PManFraig: reduced, NFramesAll: 1, PMemFraig: make([]aig.Edge, len(frames.Nodes()))}
	Fra_ObjSetFraig(p, fx, 0, rep)
	Fra_ObjSetFraig(p, fy, 0, rep.Not())
	imp := Fra_ImpCreate(int32(x.NodeID()), int32(y.NodeID()))
	vec.Vec_IntPush(b.VImps, imp)
	Fra_BmcFilterImplications(p, b)
	if vec.Vec_IntSize(b.VImps) != 1 {
		t.Fatalf("prefix-only failure removed implication: %v", vec.Vec_IntArray(b.VImps))
	}
	b.NPref = 0
	Fra_BmcFilterImplications(p, b)
	if vec.Vec_IntSize(b.VImps) != 0 {
		t.Fatalf("failure in checked frame retained implication: %v", vec.Vec_IntArray(b.VImps))
	}
}
