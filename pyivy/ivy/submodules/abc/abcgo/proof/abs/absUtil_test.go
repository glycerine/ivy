package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestAbsParametersDefaults(t *testing.T) {
	p := Abs_Par_t{nTimeOut: 99, fVerbose: 1}
	Abs_ParSetDefaults(&p)
	if p.nTimeOut != 0 || p.fVerbose != 0 || p.nFramesPast != 4 || p.nLearnedMax != 1000 || p.nLearnedDelta != 200 || p.nLearnedPerce != 70 || p.nRatioMax != 30 || p.fPropFanout != 1 || p.iFrame != -1 || p.iFrameProved != -1 || p.nFramesNoChangeLim != 2 {
		t.Fatalf("wrong defaults: %+v", p)
	}
}

func TestAbstractionFrameAndFlopConversions(t *testing.T) {
	p := gia.Gia_ManStart(20)
	pi, ro := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gate := gia.Gia_ManAppendAnd(p, pi, ro)
	gia.Gia_ManAppendCo(p, gate)
	gia.Gia_ManAppendCo(p, ro)
	gia.Gia_ManSetRegNum(p, 1)
	fla := vec.Vec_IntStart(1)
	vec.Vec_IntWriteEntry(fla, 0, 1)
	gla := Gia_FlaConvertToGla(p, fla)
	if got := Gia_GlaCountFlops(p, gla); got != 1 {
		t.Fatalf("flops=%d", got)
	}
	if got := Gia_GlaCountNodes(p, gla); got != 1 {
		t.Fatalf("nodes=%d", got)
	}
	if got := vec.Vec_IntArray(Gia_GlaConvertToFla(p, gla)); !reflect.DeepEqual(got, []int32{1}) {
		t.Fatalf("flop map=%v", got)
	}
	// A source GLA vector for frame expansion is a Boolean membership map.
	frameMap := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p)))
	vec.Vec_IntWriteEntry(frameMap, int(ro>>1), 1)
	vec.Vec_IntWriteEntry(frameMap, int(gate>>1), 1)
	vta := Gia_VtaConvertFromGla(p, frameMap, 2)
	out := Gia_VtaConvertToGla(p, vta)
	if vec.Vec_IntEntry(out, 0) != 2 || vec.Vec_IntEntry(out, int(ro>>1)) != 3 || vec.Vec_IntEntry(out, int(gate>>1)) != 3 {
		t.Fatalf("expanded GLA=%v", vec.Vec_IntArray(out))
	}
}
