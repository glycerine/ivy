package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// The AIG-facing driver must reject a non-failing trace and return a target
// that agrees with the independently known state for each extraction mode.
func TestAigTargetStatesDriver(t *testing.T) {
	g := aig.NewGraph()
	pi, reg := g.NewInput(), g.NewInput()
	g.AddOutput(reg)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	util.Abc_InfoSetBit(c.PData, 1)
	for _, mode := range []struct{ one, all int32 }{{1, 0}, {0, 0}, {0, 1}} {
		result := Bmc_AigTargetStates(g, c, 0, 1, 1, mode.one, mode.all, 0)
		if result == nil {
			t.Fatalf("no target for mode %+v", mode)
		}
		for _, input := range []bool{false, true} {
			outputs, err := result.Evaluate([]bool{input})
			if err != nil || len(outputs) != 1 || outputs[0] != input {
				t.Fatalf("mode %+v input %v got %v err %v", mode, input, outputs, err)
			}
		}
	}
	bad := util.Abc_CexAlloc(1, 1, 2)
	bad.IFrame = 1
	if Bmc_AigTargetStates(g, bad, 0, 1, 1, 1, 0, 0) != nil {
		t.Fatal("accepted a trace that does not fail")
	}
}

// Terminal priorities encode both the actual input phase and the chosen
// chronological priority; exercise both frames and both input polarities.
func TestCexPhasePriorityAcrossFrames(t *testing.T) {
	g := aig.NewGraph()
	pi, reg := g.NewInput(), g.NewInput()
	g.AddOutput(reg)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	util.Abc_InfoSetBit(c.PData, 1)
	terms := Saig_ManCexMinCollectFrameTerms(g, c)
	priorities := Saig_ManCexMinCollectPhasePriority(g, c, terms)
	if vec.Vec_VecSize(priorities) != 2 {
		t.Fatalf("frames %d", vec.Vec_VecSize(priorities))
	}
	for f := 0; f < 2; f++ {
		ids := vec.Vec_IntArray(vec.Vec_VecEntryInt(terms, f))
		vals := vec.Vec_IntArray(vec.Vec_VecEntryInt(priorities, f))
		if len(ids) != len(vals) {
			t.Fatalf("frame %d ids %v priorities %v", f, ids, vals)
		}
		for i, id := range ids {
			if id == int32(pi.NodeID()) {
				want := int32(0)
				if f == 0 {
					want = 1
				}
				if vals[i]&1 != want {
					t.Fatalf("frame %d priority %d want phase %d", f, vals[i], want)
				}
			}
		}
	}
}

func TestCexAnnotationSetAndAddIsolation(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	c := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(c.PData, 0)
	util.Abc_InfoSetBit(c.PData, 1)
	if Gia_ManAnnotateUnrolling(p, c, 0) != 1 {
		t.Fatal("annotation failed")
	}
	defer unrollingAnnotations.Delete(p)
	ao, bo := gia.Gia_ManObj(p, a>>1), gia.Gia_ManObj(p, b>>1)
	before := Gia_ManGetTwo(p, 0, bo)
	Gia_ManSetTwo(p, 0, ao, 0)
	Gia_ManAddTwo(p, 0, ao, 1)
	Gia_ManAddTwo(p, 0, ao, 2)
	if got := Gia_ManGetTwo(p, 0, ao); got != 3 {
		t.Fatalf("combined annotation %d", got)
	}
	if got := Gia_ManGetTwo(p, 0, bo); got != before {
		t.Fatalf("adjacent annotation changed: %d to %d", before, got)
	}
}

func TestRecursiveJustificationSelectsControllingInput(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gate := gia.Gia_ManAppendAnd(p, a, b)
	gia.Gia_ManAppendCo(p, gate^1)
	n := int(gia.Gia_ManObjNum(p))
	values := vec.Vec_BitStart(n)
	just := vec.Vec_BitStart(n)
	// a=false and b=true, so only a determines the false AND output.
	vec.Vec_BitWriteEntry(values, int(b>>1), 1)
	Bmc_GiaGenerateJust_rec(p, 0, gate>>1, values, just)
	if vec.Vec_BitEntry(just, int(a>>1)) != 1 || vec.Vec_BitEntry(just, int(b>>1)) != 0 {
		t.Fatalf("wrong controlling input, a=%d b=%d", vec.Vec_BitEntry(just, int(a>>1)), vec.Vec_BitEntry(just, int(b>>1)))
	}
}

func TestCexMinDiagnosticReleasesAnnotation(t *testing.T) {
	p := gia.Gia_ManStart(8)
	pi := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, pi)
	c := util.Abc_CexAlloc(0, 1, 1)
	util.Abc_InfoSetBit(c.PData, 0)
	for attempt := 0; attempt < 2; attempt++ {
		if result := Gia_ManCexMin(p, c, 0, -1, 0, 1, 0); result != nil {
			t.Fatal("diagnostic unexpectedly returned a counterexample")
		}
		if _, exists := unrollingAnnotations.Load(p); exists {
			t.Fatal("diagnostic retained annotation")
		}
	}
}

func TestCexBuildNetworkCompatibilityEntry(t *testing.T) {
	p := gia.Gia_ManStart(12)
	pi, reg := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, reg)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManSetRegNum(p, 1)
	c := util.Abc_CexAlloc(0, 2, 2)
	c.IFrame = 1
	c.PData[0] = 9
	q := Bmc_CexBuildNetwork2_(p, c, 1)
	for _, input := range []bool{false, true} {
		got, err := q.Evaluate([]bool{input})
		if err != nil || len(got) != 1 || got[0] != input {
			t.Fatalf("input %v output %v error %v", input, got, err)
		}
	}
}
