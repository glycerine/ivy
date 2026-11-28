package gia_test

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abci"
	abcmain "github.com/glycerine/abcgo/base/main"
)

// Exercise the synthesis engines and startup aliases through the real dispatcher,
// in addition to the native comparison with scripted command boundaries.
func TestDeepSynthesisRealCommandPipeline(t *testing.T) {
	previous := abcmain.Abc_FrameReadGlobalFrame()
	frame := abcmain.NewFrame()
	abcmain.Abc_FrameSetGlobalFrame(frame)
	defer abcmain.Abc_FrameSetGlobalFrame(previous)
	defer frame.Close()
	abcmain.Abc_FrameInit(frame)
	abci.Abc_Init(frame)
	gia.InstallFrameBridge(frame)
	if frame.Execute("source ../../test_vectors/abc.rc") != 0 {
		t.Fatal("loading standard synthesis aliases failed")
	}
	makeGraph := func() *gia.Gia_Man_t {
		g := gia.Gia_ManStart(32)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		for i := 0; i < 6; i++ {
			a = gia.Gia_ManAppendAnd(g, a, b)
		}
		gia.Gia_ManAppendCo(g, a)
		return g
	}
	check := func(g *gia.Gia_Man_t) {
		t.Helper()
		if g == nil || gia.Gia_ManAndNum(g) != 1 {
			t.Fatal("synthesis did not reduce redundant conjunctions to one AND")
		}
		for mask := 0; mask < 4; mask++ {
			values, err := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
			if err != nil || len(values) != 1 || values[0] != (mask == 3) {
				t.Fatalf("assignment %d: outputs %v, error %v", mask, values, err)
			}
		}
	}
	frame.SetResource("pGia", makeGraph())
	best := gia.Gia_ManDeepSynOne(0, 2, 1, 0, 0, 0)
	if best != nil {
		defer gia.Gia_ManStop(best)
	}
	check(best)
	frame.SetResource("pGia", makeGraph())
	if status := frame.Execute("&deepsyn -I 2 -J 0 -T 2 -A 1 -S 3 -t"); status != 0 {
		t.Fatalf("registered synthesis command returned %d", status)
	}
	current, ok := abcmain.Abc_FrameReadGia(frame).(*gia.Gia_Man_t)
	if !ok {
		t.Fatal("registered command did not retain a GIA graph")
	}
	check(current)
	// The direct result remains independently owned after replacing frame graphs.
	check(best)
}
