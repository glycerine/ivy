package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestAbsParametersDefaults(t *testing.T) {
	p := Abs_Par_t{NTimeOut: 99, FVerbose: 1}
	Abs_ParSetDefaults(&p)
	if p.NTimeOut != 0 || p.FVerbose != 0 || p.NFramesPast != 4 || p.nLearnedMax != 1000 || p.NLearnedDelta != 200 || p.NLearnedPerce != 70 || p.NRatioMax != 30 || p.FPropFanout != 1 || p.IFrame != -1 || p.iFrameProved != -1 || p.NFramesNoChangeLim != 2 {
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

func TestFlopToGateClassesIncludesNativeMuxControl(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	c := gia.Gia_ManAppendCi(g)
	d := gia.Gia_ManAppendCi(g)
	control := gia.Gia_ManAppendAnd(g, c, d)
	g.EnableMuxes()
	mux := gia.Gia_ManAppendMuxReal(g, control, a, b)
	gia.Gia_ManAppendCo(g, mux)
	result := Gia_FlaConvertToGla(g, vec.Vec_IntStart(0))
	if vec.Vec_IntEntry(result, int(mux>>1)) != 1 || vec.Vec_IntEntry(result, int(control>>1)) != 1 {
		t.Fatalf("MUX or control cone missing from gate classes")
	}
}

// Expected bytes were obtained by compiling the checked-in C absUtil.c and
// calling Gia_VtaConvertFromGla with nObjs=6, members 2 and 3, and 2 frames.
func TestVtaFramePackingMatchesCOracle(t *testing.T) {
	p := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, 0)
	if gia.Gia_ManObjNum(p) != 6 {
		t.Fatal("unexpected fixture size")
	}
	classes := vec.Vec_IntStart(6)
	vec.Vec_IntWriteEntry(classes, 2, 1)
	vec.Vec_IntWriteEntry(classes, 3, 1)
	got := vec.Vec_IntArray(Gia_VtaConvertFromGla(p, classes, 2))
	want := []int32{2, 4, 6, 10, 2, 3, 2, 3, 10, 11}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Go packing %v, C oracle %v", got, want)
	}
}
